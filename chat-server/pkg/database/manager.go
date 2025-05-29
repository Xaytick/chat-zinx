package database

import (
	"fmt"
	"hash/crc32"
	"math/rand"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseManager 数据库管理器
type DatabaseManager struct {
	config       *DatabaseConfig
	masterDB     *gorm.DB
	slaveDBS     []*gorm.DB
	shardDBS     map[int]*gorm.DB // 分片数据库连接
	mu           sync.RWMutex
	loadBalancer *LoadBalancer
}

// LoadBalancer 负载均衡器
type LoadBalancer struct {
	slaves []SlaveConnection
	mu     sync.RWMutex
}

// SlaveConnection 从库连接信息
type SlaveConnection struct {
	db     *gorm.DB
	weight int
	active bool
}

// NewDatabaseManager 创建新的数据库管理器
func NewDatabaseManager(config *DatabaseConfig) (*DatabaseManager, error) {
	dm := &DatabaseManager{
		config:   config,
		shardDBS: make(map[int]*gorm.DB),
		loadBalancer: &LoadBalancer{
			slaves: make([]SlaveConnection, 0),
		},
	}

	// 初始化主库连接
	if err := dm.initMaster(); err != nil {
		return nil, fmt.Errorf("failed to init master db: %w", err)
	}

	// 初始化从库连接
	if err := dm.initSlaves(); err != nil {
		return nil, fmt.Errorf("failed to init slave dbs: %w", err)
	}

	// 如果启用分片，初始化分片数据库
	if config.Sharding.Enabled {
		if err := dm.initShards(); err != nil {
			return nil, fmt.Errorf("failed to init shard dbs: %w", err)
		}
	}

	return dm, nil
}

// initMaster 初始化主库连接
func (dm *DatabaseManager) initMaster() error {
	db, err := dm.createConnection(dm.config.Master.DSN, &dm.config.Master.MaxOpenConns,
		&dm.config.Master.MaxIdleConns, &dm.config.Master.ConnMaxLifetime)
	if err != nil {
		return err
	}
	dm.masterDB = db
	return nil
}

// initSlaves 初始化从库连接
func (dm *DatabaseManager) initSlaves() error {
	for _, slaveConfig := range dm.config.Slaves {
		db, err := dm.createConnection(slaveConfig.DSN, &slaveConfig.MaxOpenConns,
			&slaveConfig.MaxIdleConns, &slaveConfig.ConnMaxLifetime)
		if err != nil {
			return err
		}

		dm.loadBalancer.mu.Lock()
		dm.loadBalancer.slaves = append(dm.loadBalancer.slaves, SlaveConnection{
			db:     db,
			weight: slaveConfig.Weight,
			active: true,
		})
		dm.loadBalancer.mu.Unlock()
	}
	return nil
}

// initShards 初始化分片数据库
func (dm *DatabaseManager) initShards() error {
	for i := 0; i < dm.config.Sharding.ShardCount; i++ {
		dsn := dm.config.GetShardDSN(i)
		db, err := dm.createConnection(dsn, &dm.config.Master.MaxOpenConns,
			&dm.config.Master.MaxIdleConns, &dm.config.Master.ConnMaxLifetime)
		if err != nil {
			return err
		}
		dm.shardDBS[i] = db
	}
	return nil
}

// createConnection 创建数据库连接
func (dm *DatabaseManager) createConnection(dsn string, maxOpen, maxIdle *int, maxLifetime *time.Duration) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if maxOpen != nil {
		sqlDB.SetMaxOpenConns(*maxOpen)
	}
	if maxIdle != nil {
		sqlDB.SetMaxIdleConns(*maxIdle)
	}
	if maxLifetime != nil {
		sqlDB.SetConnMaxLifetime(*maxLifetime)
	}

	return db, nil
}

// GetMaster 获取主库连接（用于写操作）
func (dm *DatabaseManager) GetMaster() *gorm.DB {
	return dm.masterDB
}

// GetSlave 获取从库连接（用于读操作）
func (dm *DatabaseManager) GetSlave() *gorm.DB {
	dm.loadBalancer.mu.RLock()
	defer dm.loadBalancer.mu.RUnlock()

	if len(dm.loadBalancer.slaves) == 0 {
		// 如果没有从库，使用主库
		return dm.masterDB
	}

	// 简单的随机选择算法
	index := rand.Intn(len(dm.loadBalancer.slaves))
	slave := dm.loadBalancer.slaves[index]

	if slave.active {
		return slave.db
	}

	// 如果选中的从库不可用，返回主库
	return dm.masterDB
}

// GetShardDB 获取分片数据库连接
func (dm *DatabaseManager) GetShardDB(shardKey interface{}) (*gorm.DB, error) {
	if !dm.config.Sharding.Enabled {
		return dm.masterDB, nil
	}

	shardIndex := dm.getShardIndex(shardKey)

	dm.mu.RLock()
	db, exists := dm.shardDBS[shardIndex]
	dm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("shard %d not found", shardIndex)
	}

	return db, nil
}

// getShardIndex 根据分片键计算分片索引
func (dm *DatabaseManager) getShardIndex(shardKey interface{}) int {
	var hashValue uint32

	switch v := shardKey.(type) {
	case string:
		hashValue = crc32.ChecksumIEEE([]byte(v))
	case uint:
		hashValue = uint32(v)
	case uint32:
		hashValue = v
	case int:
		hashValue = uint32(v)
	case int32:
		hashValue = uint32(v)
	default:
		hashValue = crc32.ChecksumIEEE([]byte(fmt.Sprintf("%v", v)))
	}

	return int(hashValue % uint32(dm.config.Sharding.ShardCount))
}

// GetShardTableName 获取分片表名
func (dm *DatabaseManager) GetShardTableName(tableName string, shardKey interface{}) string {
	if !dm.config.Sharding.Enabled {
		return tableName
	}

	// 查找表的分片配置
	for _, tableConfig := range dm.config.Sharding.Tables {
		if tableConfig.TableName == tableName {
			shardIndex := dm.getShardIndex(shardKey)
			return fmt.Sprintf(tableName+tableConfig.ShardSuffix, shardIndex)
		}
	}

	return tableName
}

// ExecuteInTransaction 在事务中执行操作
func (dm *DatabaseManager) ExecuteInTransaction(fn func(*gorm.DB) error) error {
	return dm.masterDB.Transaction(fn)
}

// ExecuteInShardTransaction 在分片事务中执行操作
func (dm *DatabaseManager) ExecuteInShardTransaction(shardKey interface{}, fn func(*gorm.DB) error) error {
	db, err := dm.GetShardDB(shardKey)
	if err != nil {
		return err
	}
	return db.Transaction(fn)
}

// Health 健康检查
func (dm *DatabaseManager) Health() error {
	// 检查主库
	if err := dm.checkDBHealth(dm.masterDB); err != nil {
		return fmt.Errorf("master db health check failed: %w", err)
	}

	// 检查从库
	dm.loadBalancer.mu.Lock()
	for i := range dm.loadBalancer.slaves {
		if err := dm.checkDBHealth(dm.loadBalancer.slaves[i].db); err != nil {
			dm.loadBalancer.slaves[i].active = false
		} else {
			dm.loadBalancer.slaves[i].active = true
		}
	}
	dm.loadBalancer.mu.Unlock()

	// 检查分片数据库
	for i, db := range dm.shardDBS {
		if err := dm.checkDBHealth(db); err != nil {
			return fmt.Errorf("shard %d health check failed: %w", i, err)
		}
	}

	return nil
}

// checkDBHealth 检查单个数据库健康状态
func (dm *DatabaseManager) checkDBHealth(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Close 关闭所有数据库连接
func (dm *DatabaseManager) Close() error {
	var errs []error

	// 关闭主库
	if dm.masterDB != nil {
		if sqlDB, err := dm.masterDB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	// 关闭从库
	dm.loadBalancer.mu.Lock()
	for _, slave := range dm.loadBalancer.slaves {
		if sqlDB, err := slave.db.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}
	dm.loadBalancer.mu.Unlock()

	// 关闭分片数据库
	for _, db := range dm.shardDBS {
		if sqlDB, err := db.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("multiple close errors: %v", errs)
	}

	return nil
}
