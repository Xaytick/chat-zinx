package router

import (
	"encoding/json"
	"fmt"

	"github.com/Xaytick/chat-zinx/chat-server/global"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/storage"
	"github.com/Xaytick/zinx/ziface"
	"github.com/Xaytick/zinx/znet"
)

type TextMsgRouter struct {
	znet.BaseRouter
}

func (r *TextMsgRouter) Handle(request ziface.IRequest) {
	// 1. 解析消息体
	var msg model.TextMsg
	if err := json.Unmarshal(request.GetData(), &msg); err != nil {
		// 解析失败，返回错误
		fmt.Println("消息解析失败", err)
		return
	}

	// 获取发送者信息
	fromUserID, _ := request.GetConnection().GetProperty("userID")
	fromUsername, _ := request.GetConnection().GetProperty("username")
	fmt.Printf("[消息接收] 从 %v(%v) 发送到 %s: %s\n",
		fromUsername, fromUserID, msg.ToUserID, msg.Content)

	// 2. 查找接收者用户
	// 先检查ToUserID是否为用户ID
	targetUser, err := global.UserService.GetUserByID(msg.ToUserID)
	if err != nil {
		// 不是用户ID，尝试作为用户名查找
		targetUser, err = global.UserService.GetUserByUsername(msg.ToUserID)
		if err != nil {
			// 找不到接收者，先记录日志
			fmt.Printf("[未知接收者] 用户 %s 不存在，但仍将保存为离线消息\n", msg.ToUserID)

			// 尽管找不到用户，仍然将消息存储为离线消息
			// 以便用户注册后可以收到
			storage.SaveOfflineMsg(msg.ToUserID, request.GetData())
			return
		}
	}

	// 找到了接收者，记录接收者的信息
	toUserID := targetUser.UserID
	toUsername := targetUser.Username
	fmt.Printf("[消息路由] 准备发送消息到用户: %s(ID:%s)\n", toUsername, toUserID)

	// 3. 获取全局连接管理器并寻找接收者连接
	connManager := global.GlobalServer.GetConnManager()
	fmt.Printf("[系统状态] 当前在线连接数: %d\n", connManager.Size())

	// 4. 遍历所有连接，查找目标用户
	found := false
	for _, conn := range connManager.All() {
		// 获取连接关联的用户ID
		if userIDProp, err := conn.GetProperty("userID"); err == nil {
			// 确保类型一致进行比较
			userID, ok := userIDProp.(string)
			if ok && userID == toUserID {
				fmt.Printf("[消息投递] 用户 %s 在线，直接发送消息\n", toUsername)
				// 目标用户在线，直接转发消息
				conn.SendMsg(protocol.MsgIDTextMsg, request.GetData())
				found = true
				break
			}
		}
	}

	if !found {
		fmt.Printf("[离线存储] 用户 %s 不在线，存储为离线消息\n", toUsername)
		// 5. 目标用户不在线，存储为离线消息
		storage.SaveOfflineMsg(toUserID, request.GetData())
	}
}
