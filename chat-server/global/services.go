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
)

// InitServices 初始化所有服务
func InitServices() {
	// 初始化用户服务(目前用内存实现，后续替换为数据库)
	UserService = service.NewInMemoryUserService()

	// 这里可以添加其他服务初始化
}
