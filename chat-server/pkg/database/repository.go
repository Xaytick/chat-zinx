package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// Repository 数据访问层抽象
type Repository struct {
	manager *DatabaseManager
}

// NewRepository 创建新的Repository
func NewRepository(manager *DatabaseManager) *Repository {
	return &Repository{
		manager: manager,
	}
}

// QueryOperation 查询操作类型
type QueryOperation int

const (
	ReadOperation QueryOperation = iota
	WriteOperation
	ShardReadOperation
	ShardWriteOperation
)

// QueryOptions 查询选项
type QueryOptions struct {
	Operation    QueryOperation // 操作类型
	ShardKey     interface{}    // 分片键
	TableName    string         // 表名
	ForceReplace bool           // 是否强制使用分片表名
}

// Execute 执行数据库操作
func (r *Repository) Execute(ctx context.Context, opts *QueryOptions, fn func(*gorm.DB) error) error {
	db, err := r.getDB(opts)
	if err != nil {
		return err
	}

	// 如果是分片操作，可能需要替换表名
	if opts != nil && opts.ForceReplace && opts.TableName != "" && opts.ShardKey != nil {
		shardTableName := r.manager.GetShardTableName(opts.TableName, opts.ShardKey)
		db = db.Table(shardTableName)
	}

	return fn(db.WithContext(ctx))
}

// getDB 根据操作类型获取对应的数据库连接
func (r *Repository) getDB(opts *QueryOptions) (*gorm.DB, error) {
	if opts == nil {
		return r.manager.GetMaster(), nil
	}

	switch opts.Operation {
	case ReadOperation:
		return r.manager.GetSlave(), nil
	case WriteOperation:
		return r.manager.GetMaster(), nil
	case ShardReadOperation, ShardWriteOperation:
		if opts.ShardKey == nil {
			return nil, fmt.Errorf("shard key is required for shard operations")
		}
		return r.manager.GetShardDB(opts.ShardKey)
	default:
		return r.manager.GetMaster(), nil
	}
}

// Transaction 执行事务
func (r *Repository) Transaction(ctx context.Context, opts *QueryOptions, fn func(*gorm.DB) error) error {
	if opts != nil && (opts.Operation == ShardWriteOperation || opts.Operation == ShardReadOperation) {
		if opts.ShardKey == nil {
			return fmt.Errorf("shard key is required for shard transaction")
		}
		return r.manager.ExecuteInShardTransaction(opts.ShardKey, func(tx *gorm.DB) error {
			return fn(tx.WithContext(ctx))
		})
	}

	return r.manager.ExecuteInTransaction(func(tx *gorm.DB) error {
		return fn(tx.WithContext(ctx))
	})
}

// MultiShardOperation 跨分片操作
func (r *Repository) MultiShardOperation(ctx context.Context, shardKeys []interface{}, fn func(map[interface{}]*gorm.DB) error) error {
	dbMap := make(map[interface{}]*gorm.DB)

	for _, key := range shardKeys {
		db, err := r.manager.GetShardDB(key)
		if err != nil {
			return err
		}
		dbMap[key] = db.WithContext(ctx)
	}

	return fn(dbMap)
}

// BatchInsert 批量插入（考虑分片）
func (r *Repository) BatchInsert(ctx context.Context, tableName string, items []interface{}, getShardKey func(interface{}) interface{}) error {
	if !r.manager.config.Sharding.Enabled {
		// 如果没有启用分片，直接使用主库插入
		return r.Execute(ctx, &QueryOptions{Operation: WriteOperation}, func(db *gorm.DB) error {
			return db.Table(tableName).CreateInBatches(items, 100).Error
		})
	}

	// 按分片分组
	shardGroups := make(map[int][]interface{})
	for _, item := range items {
		shardKey := getShardKey(item)
		shardIndex := r.manager.getShardIndex(shardKey)
		shardGroups[shardIndex] = append(shardGroups[shardIndex], item)
	}

	// 分别插入到对应的分片
	for shardIndex, groupItems := range shardGroups {
		db, exists := r.manager.shardDBS[shardIndex]
		if !exists {
			return fmt.Errorf("shard %d not found", shardIndex)
		}

		shardTableName := r.manager.GetShardTableName(tableName, shardIndex)
		if err := db.WithContext(ctx).Table(shardTableName).CreateInBatches(groupItems, 100).Error; err != nil {
			return fmt.Errorf("failed to insert into shard %d: %w", shardIndex, err)
		}
	}

	return nil
}

// CrossShardQuery 跨分片查询
func (r *Repository) CrossShardQuery(ctx context.Context, tableName string, fn func(*gorm.DB) *gorm.DB) ([]map[string]interface{}, error) {
	if !r.manager.config.Sharding.Enabled {
		var results []map[string]interface{}
		err := r.Execute(ctx, &QueryOptions{Operation: ReadOperation}, func(db *gorm.DB) error {
			return fn(db.Table(tableName)).Find(&results).Error
		})
		return results, err
	}

	var allResults []map[string]interface{}

	// 在所有分片上执行查询
	for i := 0; i < r.manager.config.Sharding.ShardCount; i++ {
		db, exists := r.manager.shardDBS[i]
		if !exists {
			continue
		}

		shardTableName := r.manager.GetShardTableName(tableName, i)
		var shardResults []map[string]interface{}

		if err := fn(db.WithContext(ctx).Table(shardTableName)).Find(&shardResults).Error; err != nil {
			return nil, fmt.Errorf("failed to query shard %d: %w", i, err)
		}

		allResults = append(allResults, shardResults...)
	}

	return allResults, nil
}

// GetMaster 获取主库连接
func (r *Repository) GetMaster() *gorm.DB {
	return r.manager.GetMaster()
}

// GetSlave 获取从库连接
func (r *Repository) GetSlave() *gorm.DB {
	return r.manager.GetSlave()
}

// GetShardDB 获取分片数据库连接
func (r *Repository) GetShardDB(shardKey interface{}) (*gorm.DB, error) {
	return r.manager.GetShardDB(shardKey)
}
