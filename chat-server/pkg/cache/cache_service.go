package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	redisDao "github.com/Xaytick/chat-zinx/chat-server/dao/redis"
	"github.com/go-redis/redis/v8"
)

// CacheService 缓存服务接口
type CacheService interface {
	// 基础操作
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string) (string, error)
	Del(key string) error
	Exists(key string) (int64, error)

	// 用户会话缓存
	SetUserSession(userID string, sessionData interface{}, ttl time.Duration) error
	GetUserSession(userID string) (string, error)
	DelUserSession(userID string) error

	// 用户在线状态
	SetUserOnline(userID string, serverID string) error
	GetUserServer(userID string) (string, error)
	SetUserOffline(userID string) error

	// 消息缓存
	CacheMessage(messageID string, message interface{}, ttl time.Duration) error
	GetCachedMessage(messageID string) (string, error)

	// 群组信息缓存
	CacheGroupInfo(groupID string, groupInfo interface{}, ttl time.Duration) error
	GetCachedGroupInfo(groupID string) (string, error)

	// 消息推送队列
	PushMessage(queue string, message interface{}) error
	PopMessage(queue string) (string, error)

	// 健康检查
	Health() error
}

// UnifiedCacheService 统一缓存服务实现
type UnifiedCacheService struct {
	client redis.UniversalClient
}

// NewCacheService 创建缓存服务实例
func NewCacheService() CacheService {
	return &UnifiedCacheService{
		client: redisDao.GetUniversalClient(),
	}
}

// Set 设置键值对
func (c *UnifiedCacheService) Set(key string, value interface{}, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func (c *UnifiedCacheService) Get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.client.Get(ctx, key).Result()
}

// Del 删除键
func (c *UnifiedCacheService) Del(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.client.Del(ctx, key).Err()
}

// Exists 检查键是否存在
func (c *UnifiedCacheService) Exists(key string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.client.Exists(ctx, key).Result()
}

// SetUserSession 设置用户会话
func (c *UnifiedCacheService) SetUserSession(userID string, sessionData interface{}, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("user:session:%s", userID)
	data, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

// GetUserSession 获取用户会话
func (c *UnifiedCacheService) GetUserSession(userID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("user:session:%s", userID)
	return c.client.Get(ctx, key).Result()
}

// DelUserSession 删除用户会话
func (c *UnifiedCacheService) DelUserSession(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("user:session:%s", userID)
	return c.client.Del(ctx, key).Err()
}

// SetUserOnline 设置用户在线状态
func (c *UnifiedCacheService) SetUserOnline(userID string, serverID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("user:online:%s", userID)
	return c.client.Set(ctx, key, serverID, 24*time.Hour).Err()
}

// GetUserServer 获取用户所在服务器
func (c *UnifiedCacheService) GetUserServer(userID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("user:online:%s", userID)
	return c.client.Get(ctx, key).Result()
}

// SetUserOffline 设置用户离线
func (c *UnifiedCacheService) SetUserOffline(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("user:online:%s", userID)
	return c.client.Del(ctx, key).Err()
}

// CacheMessage 缓存消息
func (c *UnifiedCacheService) CacheMessage(messageID string, message interface{}, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("message:cache:%s", messageID)
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedMessage 获取缓存的消息
func (c *UnifiedCacheService) GetCachedMessage(messageID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("message:cache:%s", messageID)
	return c.client.Get(ctx, key).Result()
}

// CacheGroupInfo 缓存群组信息
func (c *UnifiedCacheService) CacheGroupInfo(groupID string, groupInfo interface{}, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("group:info:%s", groupID)
	data, err := json.Marshal(groupInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal group info: %w", err)
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedGroupInfo 获取缓存的群组信息
func (c *UnifiedCacheService) GetCachedGroupInfo(groupID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("group:info:%s", groupID)
	return c.client.Get(ctx, key).Result()
}

// PushMessage 推送消息到队列
func (c *UnifiedCacheService) PushMessage(queue string, message interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return c.client.LPush(ctx, queue, data).Err()
}

// PopMessage 从队列弹出消息
func (c *UnifiedCacheService) PopMessage(queue string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.client.RPop(ctx, queue).Result()
}

// Health 健康检查
func (c *UnifiedCacheService) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.client.Ping(ctx).Err()
}
