package router

import (
	"fmt"

	"github.com/Xaytick/zinx/utils"
	"github.com/Xaytick/zinx/ziface"
	"github.com/Xaytick/zinx/znet"
)

// HeartbeatRouter 心跳消息处理路由
type HeartbeatRouter struct {
	znet.BaseRouter
}

// Handle 处理心跳消息
func (hr *HeartbeatRouter) Handle(request ziface.IRequest) {
	// 获取连接
	conn := request.GetConnection()

	// 获取用户信息（如果已登录）
	userID, err := conn.GetProperty("userID")
	username := "未登录用户"
	if err == nil {
		usernameObj, _ := conn.GetProperty("username")
		if usernameObj != nil {
			username = usernameObj.(string)
		}
		// Safely assert and print userID as uint
		if uid, ok := userID.(uint); ok {
			fmt.Printf("收到用户 %s(ID=%d) 的心跳请求\n", username, uid)
		} else {
			// Fallback or log error if type is not uint as expected
			fmt.Printf("收到用户 %s(ID=%v, type error) 的心跳请求\n", username, userID)
		}
	} else {
		fmt.Printf("收到未登录连接 (ConnID=%d) 的心跳请求\n", conn.GetConnID())
	}

	// 回复一个PONG消息
	err = conn.SendMsg(utils.PONG_MSG_ID, []byte("pong"))
	if err != nil {
		fmt.Println("回复心跳消息失败:", err)
	} else {
		fmt.Printf("心跳响应: PONG 发送至 %s(ConnID=%d)\n", username, conn.GetConnID())
	}
}
