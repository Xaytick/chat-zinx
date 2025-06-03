package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/Xaytick/chat-zinx/chat-server/conf"
	"github.com/go-redis/redis/v8"
)

var (
	RedisClient        *redis.Client          // Redis单机客户端实例
	RedisClusterClient *redis.ClusterClient   // Redis集群客户端实例
	Ctx                = context.Background() // 上下文
	IsClusterMode      bool                   // 是否为集群模式
)

// InitRedis 初始化Redis连接
func InitRedis(cfg *conf.RedisConfig) error {
	if cfg.ClusterEnabled && len(cfg.ClusterAddrs) > 0 {
		// 集群模式
		return initRedisCluster(cfg)
	} else {
		// 单机模式
		return initRedisSingle(cfg)
	}
}

// initRedisSingle 初始化Redis单机连接
func initRedisSingle(cfg *conf.RedisConfig) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	if _, err := RedisClient.Ping(Ctx).Result(); err != nil {
		return fmt.Errorf("Redis单机连接失败: %w", err)
	}

	IsClusterMode = false
	fmt.Println("Redis单机模式连接成功")
	return nil
}

// initRedisCluster 初始化Redis集群连接
func initRedisCluster(cfg *conf.RedisConfig) error {
	// 设置默认值
	poolSize := cfg.PoolSize
	if poolSize <= 0 {
		poolSize = 10
	}

	minIdleConns := cfg.MinIdleConns
	if minIdleConns <= 0 {
		minIdleConns = 5
	}

	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	RedisClusterClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        cfg.ClusterAddrs,
		Password:     cfg.Password,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		MaxRetries:   maxRetries,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// 测试连接
	if err := RedisClusterClient.Ping(Ctx).Err(); err != nil {
		return fmt.Errorf("Redis集群连接失败: %w", err)
	}

	IsClusterMode = true
	fmt.Println("Redis集群模式连接成功")
	return nil
}

// GetRedisClient 获取Redis客户端
func GetRedisClient() *redis.Client {
	return RedisClient
}

// GetRedisClusterClient 获取Redis集群客户端
func GetRedisClusterClient() *redis.ClusterClient {
	return RedisClusterClient
}

// GetUniversalClient 获取通用客户端接口
func GetUniversalClient() redis.UniversalClient {
	if IsClusterMode {
		return RedisClusterClient
	}
	return RedisClient
}

// GetMessageExpiration 获取消息过期时间
func GetMessageExpiration(cfg *conf.RedisConfig) time.Duration {
	return time.Duration(cfg.MessageExpiration) * time.Second
}

// IsClusterEnabled 检查是否启用集群模式
func IsClusterEnabled() bool {
	return IsClusterMode
}

// Close 关闭Redis连接
func Close() error {
	if IsClusterMode && RedisClusterClient != nil {
		return RedisClusterClient.Close()
	} else if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}
