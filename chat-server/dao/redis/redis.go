package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/Xaytick/chat-zinx/chat-server/conf"
	"github.com/go-redis/redis/v8"
)

var (
	RedisClient *redis.Client // Redis客户端实例
	Ctx = context.Background() // 上下文
)

// InitRedis 初始化Redis连接
func InitRedis(cfg *conf.RedisConfig) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB: cfg.DB,
	})	

	// 测试连接
	if _, err := RedisClient.Ping(Ctx).Result(); err != nil {
		return err
	}

	return nil
}

// GetRedisClient 获取Redis客户端
func GetRedisClient() *redis.Client {
	return RedisClient
}

// GetMessageExpiration 获取消息过期时间
func GetMessageExpiration(cfg *conf.RedisConfig) time.Duration {
	return time.Duration(cfg.MessageExpiration) * time.Second
}



