package router

import (
	"encoding/json"
	"fmt"

	"github.com/Xaytick/chat-zinx/chat-server/global"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
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
	fromUserID, err := request.GetConnection().GetProperty("userID")
	if err != nil {
		fmt.Println("获取发送者ID失败", err)
		return
	}

	fromUsername, _ := request.GetConnection().GetProperty("username")
	fmt.Printf("[消息接收] 从 %v(%v) 发送到 %s: %s\n",
		fromUsername, fromUserID, msg.ToUserID, msg.Content)

	// 设置发送者ID
	msg.FromUserID = fromUserID.(string)

	// 重新序列化消息，包含发送者ID
	msgData, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("消息序列化失败", err)
		return
	}

	// 2. 查找接收者用户
	// 先检查ToUserID是否为用户ID
	var targetUser *model.User
	targetUser, err = global.UserService.GetUserByID(msg.ToUserID)
	if err != nil {
		// 不是用户ID，尝试作为用户名查找
		targetUser, err = global.UserService.GetUserByUsername(msg.ToUserID)
		if err != nil {
			// 找不到接收者，先记录日志
			fmt.Printf("[未知接收者] 用户 %s 不存在，但仍将保存为离线消息\n", msg.ToUserID)

			// 尽管找不到用户，仍然将消息存储为离线消息，以便用户注册后可以收到
			// 使用用户名作为临时ID
			global.MessageService.SaveMessage(fromUserID.(string), msg.ToUserID, msgData)
			return
		}
	}

	// 找到了接收者，记录接收者的信息
	toUserID := targetUser.UserID
	toUsername := targetUser.Username
	fmt.Printf("[消息路由] 准备发送消息到用户: %s(ID:%s)\n", toUsername, toUserID)

	// 4. 获取全局连接管理器并寻找接收者连接
	connManager := global.GlobalServer.GetConnManager()
	fmt.Printf("[系统状态] 当前在线连接数: %d\n", connManager.Size())

	// 5. 遍历所有连接，查找目标用户
	found := false
	for _, conn := range connManager.All() {
		// 获取连接关联的用户ID
		if userIDProp, err := conn.GetProperty("userID"); err == nil {
			// 确保类型一致进行比较
			userID, ok := userIDProp.(string)
			if ok && userID == toUserID {
				fmt.Printf("[消息投递] 用户 %s 在线，直接发送消息\n", toUsername)
				// 目标用户在线，直接转发消息
				err := conn.SendMsg(protocol.MsgIDTextMsg, msgData)
				if err != nil {
					fmt.Printf("[消息发送] 发送失败: %v, 将保存为离线消息\n", err)
				} else {
					found = true
					break
				}
			}
		}
	}

	if !found {
		fmt.Printf("[离线存储] 用户 %s 不在线或消息发送失败，存储为离线消息\n", toUsername)
		// 3. 只有当用户不在线或消息发送失败时，才保存到Redis（历史记录和离线消息）
		err = global.MessageService.SaveMessage(fromUserID.(string), toUserID, msgData)
		if err != nil {
			fmt.Printf("[Redis消息] 保存消息失败: %v\n", err)
			return
		}
	} else {
		// 用户在线并且消息发送成功，只保存历史记录不保存离线消息
		err = global.MessageService.SaveHistoryOnly(fromUserID.(string), toUserID, msgData)
		if err != nil {
			fmt.Printf("[Redis消息] 保存历史记录失败: %v\n", err)
			return
		}
		fmt.Printf("[历史记录] 用户 %s 在线且消息发送成功，只记录历史不保存离线消息\n", toUsername)
	}
}
