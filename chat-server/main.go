package main

import (
	"fmt"

	"github.com/Xaytick/chat-zinx/chat-server/global"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
	"github.com/Xaytick/chat-zinx/chat-server/router"
	"github.com/Xaytick/zinx/ziface"
	"github.com/Xaytick/zinx/znet"
)

func main() {
	// 1. 初始化所有服务
	fmt.Println("初始化服务...")
	global.InitServices()

	// 2. 创建 Zinx TCP Server
	fmt.Println("创建服务器...")
	global.GlobalServer = znet.NewServer("ChatServer")

	// 3. 注册业务路由
	fmt.Println("注册路由...")
	// 注册/登录路由
	global.GlobalServer.AddRouter(protocol.MsgIDRegisterReq, &router.RegisterRouter{})
	global.GlobalServer.AddRouter(protocol.MsgIDLoginReq, &router.LoginRouter{})
	// 聊天相关路由
	global.GlobalServer.AddRouter(protocol.MsgIDTextMsg, &router.TextMsgRouter{})

	// 4. 启动服务器勾子
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

	// 5. 启动并阻塞服务
	fmt.Println("启动服务器...")
	global.GlobalServer.Serve()
}
