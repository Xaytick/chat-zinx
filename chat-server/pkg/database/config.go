package database

import (
	"fmt"
	"time"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Master     MasterConfig   `json:"master"`      // 主库配置
	Slaves     []SlaveConfig  `json:"slaves"`      // 从库配置
	Sharding   ShardingConfig `json:"sharding"`    // 分片配置
	MaxRetries int            `json:"max_retries"` // 最大重试次数
	RetryDelay time.Duration  `json:"retry_delay"` // 重试延迟
}

// MasterConfig 主库配置
type MasterConfig struct {
	DSN             string        `json:"dsn"`               // 数据源名称
	MaxOpenConns    int           `json:"max_open_conns"`    // 最大打开连接数
	MaxIdleConns    int           `json:"max_idle_conns"`    // 最大空闲连接数
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"` // 连接最大生存时间
}

// SlaveConfig 从库配置
type SlaveConfig struct {
	DSN             string        `json:"dsn"`               // 数据源名称
	Weight          int           `json:"weight"`            // 权重（用于负载均衡）
	MaxOpenConns    int           `json:"max_open_conns"`    // 最大打开连接数
	MaxIdleConns    int           `json:"max_idle_conns"`    // 最大空闲连接数
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"` // 连接最大生存时间
}

// ShardingConfig 分片配置
type ShardingConfig struct {
	Enabled    bool               `json:"enabled"`     // 是否启用分片
	ShardCount int                `json:"shard_count"` // 分片数量
	Strategy   string             `json:"strategy"`    // 分片策略 (user_id, hash, range)
	Tables     []ShardTableConfig `json:"tables"`      // 需要分片的表配置
}

// ShardTableConfig 分片表配置
type ShardTableConfig struct {
	TableName   string `json:"table_name"`   // 表名
	ShardKey    string `json:"shard_key"`    // 分片键
	ShardSuffix string `json:"shard_suffix"` // 分片后缀格式，如 "_%02d"
}

// GetShardDSN 获取分片数据库的DSN
func (c *DatabaseConfig) GetShardDSN(shardIndex int) string {
	// 这里假设每个分片都有独立的数据库
	// 实际项目中可能需要根据具体需求调整
	baseDSN := c.Master.DSN
	return fmt.Sprintf("%s_shard_%02d", baseDSN, shardIndex)
}

// GetSlaveByWeight 根据权重选择从库
func (c *DatabaseConfig) GetSlaveByWeight() *SlaveConfig {
	if len(c.Slaves) == 0 {
		return nil
	}

	// 简单的权重轮询算法
	totalWeight := 0
	for _, slave := range c.Slaves {
		totalWeight += slave.Weight
	}

	if totalWeight == 0 {
		return &c.Slaves[0]
	}

	// 这里可以实现更复杂的负载均衡算法
	// 为简化，返回第一个从库
	return &c.Slaves[0]
}
