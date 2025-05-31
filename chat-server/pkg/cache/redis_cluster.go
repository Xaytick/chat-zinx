package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisClusterConfig Redis集群配置
type RedisClusterConfig struct {
	Addrs        []string      `json:"addrs"`          // 集群节点地址
	Password     string        `json:"password"`       // 密码
	PoolSize     int           `json:"pool_size"`      // 连接池大小
	MinIdleConns int           `json:"min_idle_conns"` // 最小空闲连接
	MaxRetries   int           `json:"max_retries"`    // 最大重试次数
	DialTimeout  time.Duration `json:"dial_timeout"`   // 连接超时
	ReadTimeout  time.Duration `json:"read_timeout"`   // 读取超时
	WriteTimeout time.Duration `json:"write_timeout"`  // 写入超时
}

// RedisClusterManager Redis集群管理器
type RedisClusterManager struct {
	client *redis.ClusterClient
	config *RedisClusterConfig
}

// NewRedisClusterManager 创建Redis集群管理器
func NewRedisClusterManager(config *RedisClusterConfig) (*RedisClusterManager, error) {
	if len(config.Addrs) == 0 {
		return nil, fmt.Errorf("Redis cluster addrs cannot be empty")
	}

	// 设置默认值
	if config.PoolSize <= 0 {
		config.PoolSize = 10
	}
	if config.MinIdleConns <= 0 {
		config.MinIdleConns = 5
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}
	if config.DialTimeout <= 0 {
		config.DialTimeout = 5 * time.Second
	}
	if config.ReadTimeout <= 0 {
		config.ReadTimeout = 3 * time.Second
	}
	if config.WriteTimeout <= 0 {
		config.WriteTimeout = 3 * time.Second
	}

	// 创建集群客户端
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        config.Addrs,
		Password:     config.Password,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis cluster: %w", err)
	}

	log.Println("Redis Cluster connected successfully")

	return &RedisClusterManager{
		client: client,
		config: config,
	}, nil
}

// GetClient 获取Redis集群客户端
func (r *RedisClusterManager) GetClient() *redis.ClusterClient {
	return r.client
}

// Close 关闭Redis集群连接
func (r *RedisClusterManager) Close() error {
	return r.client.Close()
}

// Health 健康检查
func (r *RedisClusterManager) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return r.client.Ping(ctx).Err()
}

// GetClusterInfo 获取集群信息
func (r *RedisClusterManager) GetClusterInfo() (*ClusterInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 获取集群节点信息
	nodes, err := r.client.ClusterNodes(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster nodes: %w", err)
	}

	// 获取集群插槽信息
	slots, err := r.client.ClusterSlots(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster slots: %w", err)
	}

	return &ClusterInfo{
		Nodes: nodes,
		Slots: fmt.Sprintf("%+v", slots),
	}, nil
}

// ClusterInfo 集群信息
type ClusterInfo struct {
	Nodes string `json:"nodes"`
	Slots string `json:"slots"`
}

// ChatCache 聊天缓存接口
type ChatCache interface {
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
}

// ChatCacheManager 聊天缓存管理器
type ChatCacheManager struct {
	cluster *RedisClusterManager
}

// NewChatCacheManager 创建聊天缓存管理器
func NewChatCacheManager(cluster *RedisClusterManager) *ChatCacheManager {
	return &ChatCacheManager{
		cluster: cluster,
	}
}

// SetUserSession 设置用户会话
func (c *ChatCacheManager) SetUserSession(userID string, sessionData interface{}, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("user:session:%s", userID)
	data, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	return c.cluster.client.Set(ctx, key, data, ttl).Err()
}

// GetUserSession 获取用户会话
func (c *ChatCacheManager) GetUserSession(userID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("user:session:%s", userID)
	return c.cluster.client.Get(ctx, key).Result()
}

// DelUserSession 删除用户会话
func (c *ChatCacheManager) DelUserSession(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("user:session:%s", userID)
	return c.cluster.client.Del(ctx, key).Err()
}

// SetUserOnline 设置用户在线状态
func (c *ChatCacheManager) SetUserOnline(userID string, serverID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("user:online:%s", userID)
	return c.cluster.client.Set(ctx, key, serverID, 30*time.Minute).Err()
}

// GetUserServer 获取用户所在服务器
func (c *ChatCacheManager) GetUserServer(userID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("user:online:%s", userID)
	return c.cluster.client.Get(ctx, key).Result()
}

// SetUserOffline 设置用户离线
func (c *ChatCacheManager) SetUserOffline(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("user:online:%s", userID)
	return c.cluster.client.Del(ctx, key).Err()
}

// CacheMessage 缓存消息
func (c *ChatCacheManager) CacheMessage(messageID string, message interface{}, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("message:%s", messageID)
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return c.cluster.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedMessage 获取缓存消息
func (c *ChatCacheManager) GetCachedMessage(messageID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("message:%s", messageID)
	return c.cluster.client.Get(ctx, key).Result()
}

// CacheGroupInfo 缓存群组信息
func (c *ChatCacheManager) CacheGroupInfo(groupID string, groupInfo interface{}, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("group:info:%s", groupID)
	data, err := json.Marshal(groupInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal group info: %w", err)
	}

	return c.cluster.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedGroupInfo 获取缓存群组信息
func (c *ChatCacheManager) GetCachedGroupInfo(groupID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("group:info:%s", groupID)
	return c.cluster.client.Get(ctx, key).Result()
}

// PushMessage 推送消息到队列
func (c *ChatCacheManager) PushMessage(queue string, message interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return c.cluster.client.LPush(ctx, queue, data).Err()
}

// PopMessage 从队列弹出消息
func (c *ChatCacheManager) PopMessage(queue string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.cluster.client.RPop(ctx, queue).Result()
}
