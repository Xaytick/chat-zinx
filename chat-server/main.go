package main

import (
	"github.com/Xaytick/chat-zinx/chat-server/global"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
	"github.com/Xaytick/chat-zinx/chat-server/router"
	"github.com/Xaytick/zinx/znet"
)

func main() {
	// 1. 创建 Zinx TCP Server
	global.GlobalServer = znet.NewServer("ChatServer")

	// 2. TODO: 在此处调用 server.AddRouter(msgID, router) 注册业务路由
	global.GlobalServer.AddRouter(protocol.MsgIDLoginReq, &router.LoginRouter{})
	global.GlobalServer.AddRouter(protocol.MsgIDTextMsg, &router.TextMsgRouter{})

	// 3. 启动并阻塞服务
	global.GlobalServer.Serve()
}
