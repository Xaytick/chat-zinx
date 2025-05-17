package main

import (
	"fmt"
	"log"

	"github.com/Xaytick/chat-zinx/chat-server/conf"
	"github.com/Xaytick/chat-zinx/chat-server/dao/mysql"
	"github.com/Xaytick/chat-zinx/chat-server/dao/redis"
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

	// 初始化Redis连接
	fmt.Println("初始化Redis连接...")
	err = redis.InitRedis(conf.GetRedisConfig())
	if err != nil {
		log.Fatalf("初始化Redis失败: %v", err)
	}
	fmt.Println("Redis连接成功!")

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

	// 聊天消息路由
	global.GlobalServer.AddRouter(protocol.MsgIDTextMsg, &router.TextMsgRouter{})

	// 历史消息和聊天关系路由
	global.GlobalServer.AddRouter(protocol.MsgIDHistoryMsgReq, &router.HistoryMsgRouter{})
	global.GlobalServer.AddRouter(protocol.MsgIDChatRelationReq, &router.ChatRelationRouter{})

	// 群组功能路由 (Group feature routers)
	global.GlobalServer.AddRouter(protocol.MsgIDCreateGroupReq, &router.CreateGroupRouter{})
	global.GlobalServer.AddRouter(protocol.MsgIDJoinGroupReq, &router.JoinGroupRouter{})
	global.GlobalServer.AddRouter(protocol.MsgIDLeaveGroupReq, &router.LeaveGroupRouter{})

	// 6. 设置心跳检测，启用心跳检测会自动启动心跳路由
	fmt.Println("启用心跳检测...")
	global.GlobalServer.SetHeartbeat(true)

	// 7. 启动服务器勾子
	// 设置连接开始时的钩子函数
	global.GlobalServer.SetOnConnStart(func(conn ziface.IConnection) {
		fmt.Println("新连接 ConnID=", conn.GetConnID(), "IP:", conn.RemoteAddr().String())
		fmt.Printf("心跳检测已启用，间隔: %d秒，超时: %d秒\n",
			conf.GetHeartbeatInterval(), conf.GetHeartbeatTimeout())
	})

	// 设置连接结束时的钩子函数
	global.GlobalServer.SetOnConnStop(func(conn ziface.IConnection) {
		if userID, err := conn.GetProperty("userID"); err == nil {
			username, _ := conn.GetProperty("username")
			fmt.Printf("连接断开 ConnID=%d, 用户: %s(ID=%s)\n",
				conn.GetConnID(), username, userID)
		} else {
			fmt.Println("连接断开 ConnID=", conn.GetConnID(), "未登录用户")
		}
	})

	// 8. 启动并阻塞服务
	fmt.Println("启动服务器...")
	global.GlobalServer.Serve()
}
