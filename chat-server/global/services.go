package global

import (
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
)

// InitServices 初始化所有服务
func InitServices() {
	// 初始化用户服务(使用MySQL实现)
	UserService = service.NewMySQLUserService()

	// 初始化消息服务(使用Redis实现)
	MessageService = service.NewRedisMessageService()

	// 这里可以添加其他服务初始化
}
 