package main

import (
	"fmt"
	"log"

	"github.com/Xaytick/chat-zinx/chat-server/conf"
	"github.com/Xaytick/chat-zinx/chat-server/dao/mysql"
	"github.com/Xaytick/chat-zinx/chat-server/global"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
	"github.com/Xaytick/chat-zinx/chat-server/router"
	"github.com/Xaytick/zinx/ziface"
	"github.com/Xaytick/zinx/znet"
)

func main() {
	// 1. 加载配置
	fmt.Println("加载配置...")
	config, err := conf.LoadConfig("./conf/config.json")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	fmt.Println("配置加载成功!")

	// 2. 初始化数据库连接
	fmt.Println("初始化MySQL连接...")
	err = mysql.InitMySQL(conf.GetMySQLConfig())
	if err != nil {
		log.Fatalf("初始化MySQL失败: %v", err)
	}
	fmt.Println("MySQL连接成功!")

	// 3. 初始化所有服务
	fmt.Println("初始化服务...")
	global.InitServices()

	// 4. 创建 Zinx TCP Server
	fmt.Println("创建服务器...")
	global.GlobalServer = znet.NewServer(config.Name)

	// 5. 注册业务路由
	fmt.Println("注册路由...")
	// 注册/登录路由
	global.GlobalServer.AddRouter(protocol.MsgIDRegisterReq, &router.RegisterRouter{})
	global.GlobalServer.AddRouter(protocol.MsgIDLoginReq, &router.LoginRouter{})
	// 聊天相关路由
	global.GlobalServer.AddRouter(protocol.MsgIDTextMsg, &router.TextMsgRouter{})

	// 6. 启动服务器勾子
	// 设置连接开始时的钩子函数
	global.GlobalServer.SetOnConnStart(func(conn ziface.IConnection) {
		fmt.Println("新连接 ConnID=", conn.GetConnID(), "IP:", conn.RemoteAddr().String())
	})

	// 设置连接结束时的钩子函数
	global.GlobalServer.SetOnConnStop(func(conn ziface.IConnection) {
		if userID, err := conn.GetProperty("userID"); err == nil {
			fmt.Println("连接断开 ConnID=", conn.GetConnID(), "用户ID:", userID)
		} else {
			fmt.Println("连接断开 ConnID=", conn.GetConnID(), "未登录用户")
		}
	})

	// 7. 启动并阻塞服务
	fmt.Println("启动服务器...")
	global.GlobalServer.Serve()
}
