package examples

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Xaytick/chat-zinx/chat-server/conf"
	"github.com/Xaytick/chat-zinx/chat-server/dao/redis"
	"github.com/Xaytick/zinx/znet"
)

// InitAuthServer 初始化带认证的服务器
func InitAuthServer() *znet.Server {
	// 1. 加载配置
	configPath := filepath.Join("conf", "config.json")
	config, err := conf.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化Redis（用于存储nonce和会话）
	redisConfig := conf.GetRedisConfig()
	err = redis.InitRedis(redisConfig)
	if err != nil {
		fmt.Printf("初始化Redis失败: %v\n", err)
		os.Exit(1)
	}

	// 3. 创建服务器
	// 使用配置中的名称创建服务器
	server := znet.NewServer(config.Name)

	// 4. 注册路由（使用带认证的路由）
	SetupAuthRouters(server)

	return server
}

// 在main.go中使用：
/*
func main() {
	// 初始化带认证的服务器
	server := examples.InitAuthServer()

	// 启动服务器
	server.Serve()
}
*/
