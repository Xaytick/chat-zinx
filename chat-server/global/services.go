package global

import (
	"fmt"

	"github.com/Xaytick/chat-zinx/chat-server/pkg/cache"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/service"
	"github.com/Xaytick/zinx/ziface"
)

// Config 配置结构体
type AppConfig struct {
	IsDev      bool   `json:"is_dev"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	NatsURLs   string `json:"nats_urls"`
	ConsulAddr string `json:"consul_addr"`
}

var (
	// GlobalServer 全局服务器实例
	GlobalServer ziface.IServer

	// DistributedManager 分布式管理器实例 (使用interface{}避免循环引用)
	DistributedManager interface{}

	// UserService 用户服务实例
	UserService service.IUserService

	// MessageService 消息服务实例
	MessageService service.IMessageService

	// GroupService 群组服务实例
	GroupService service.IGroupService

	// CacheService 缓存服务实例
	CacheService cache.CacheService

	// Config 应用配置
	Config *AppConfig
)

// InitServices 初始化所有服务
func InitServices() {
	// 初始化配置
	if Config == nil {
		Config = &AppConfig{
			IsDev:      false,
			Host:       "localhost",
			Port:       9000,
			NatsURLs:   "nats://localhost:4222",
			ConsulAddr: "localhost:8500",
		}
	}

	// 初始化缓存服务
	CacheService = cache.NewCacheService()

	// 初始化用户服务(使用MySQL实现)
	UserService = service.NewMySQLUserService()

	// 初始化消息服务(使用Redis实现)
	MessageService = service.NewRedisMessageService()

	// 初始化群组服务
	GroupService = service.NewGroupService()

	fmt.Println("所有服务初始化完毕!")
}
