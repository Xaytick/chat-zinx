package global

import (
	"fmt"

	"github.com/Xaytick/chat-zinx/chat-server/pkg/cache"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/service"
	"github.com/Xaytick/zinx/ziface"
)

var (
	// GlobalServer 全局服务器实例
	GlobalServer ziface.IServer

	// UserService 用户服务实例
	UserService service.IUserService

	// MessageService 消息服务实例
	MessageService service.IMessageService

	// GroupService 群组服务实例
	GroupService service.IGroupService

	// CacheService 缓存服务实例
	CacheService cache.CacheService
)

// InitServices 初始化所有服务
func InitServices() {
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
