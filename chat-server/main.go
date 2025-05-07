package main

import (
	"github.com/Xaytick/chat-server/pkg/protocol"
	"github.com/Xaytick/chat-server/router"
	"github.com/Xaytick/zinx/znet"
)

func main() {
	// 1. 创建 Zinx TCP Server
	server := znet.NewServer("ChatServer")

	// 2. TODO: 在此处调用 server.AddRouter(msgID, router) 注册业务路由
	server.AddRouter(protocol.MsgIDLoginReq, &router.LoginRouter{})

	// 3. 启动并阻塞服务
	server.Serve()
}
