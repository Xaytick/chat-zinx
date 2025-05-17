package router

import (
	"encoding/json"
	"fmt"
	"strconv"

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
		fmt.Println("消息解析失败", err)
		// TODO: Send error response to client
		return
	}

	// 获取发送者信息
	fromUserIDProp, err := request.GetConnection().GetProperty("userID")
	if err != nil {
		fmt.Println("获取发送者ID失败", err)
		// TODO: Send error response to client or close connection
		return
	}
	fromUserIDUint, ok := fromUserIDProp.(uint)
	if !ok {
		fmt.Println("发送者ID类型错误", err)
		// TODO: Send error response to client or close connection
		return
	}

	fromUserUUIDProp, _ := request.GetConnection().GetProperty("userUUID") // Assuming userUUID is also set
	fromUserUUIDStr, _ := fromUserUUIDProp.(string)

	fromUsernameProp, _ := request.GetConnection().GetProperty("username")
	fromUsernameStr, _ := fromUsernameProp.(string)

	fmt.Printf("[消息接收] 从 %s (ID: %d, UUID: %s) 发送到 %s: %s\n",
		fromUsernameStr, fromUserIDUint, fromUserUUIDStr, msg.ToUserID, msg.Content)

	// 设置发送者ID (TextMsg model uses string for FromUserID for client compatibility)
	msg.FromUserID = fromUserUUIDStr

	// 重新序列化消息，包含发送者ID (as string)
	msgData, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("消息序列化失败", err)
		// TODO: Send error response to client
		return
	}

	// 2. 查找接收者用户
	var targetUser *model.User
	// Try to find user by UUID first (assuming msg.ToUserID could be a UUID)
	// This requires UserService to have GetUserByUUID method
	targetUser, err = global.UserService.GetUserByUUID(msg.ToUserID)
	if err != nil {
		// If not found by UUID, try by username
		targetUser, err = global.UserService.GetUserByUsername(msg.ToUserID)
		if err != nil {
			// If still not found, try if msg.ToUserID is a numeric ID string
			parsedToID, parseErr := strconv.ParseUint(msg.ToUserID, 10, 32)
			if parseErr == nil {
				targetUser, err = global.UserService.GetUserByID(uint(parsedToID))
			}
		}
	}

	if err != nil || targetUser == nil {
		fmt.Printf("[未知接收者] 用户 %s 查找失败: %v. 消息不会发送.\n", msg.ToUserID, err)
		// TODO: Send "user not found" response to client
		// Example: sendResponse(request, protocol.MsgIDTextMsgResp, 1, "接收用户不存在", nil)
		return
	}

	// 找到了接收者
	toUserIDUint := targetUser.ID        // This is uint
	toUserUUIDStr := targetUser.UserUUID // This is string
	toUsernameStr := targetUser.Username

	fmt.Printf("[消息路由] 准备发送消息到用户: %s (ID: %d, UUID: %s)\n", toUsernameStr, toUserIDUint, toUserUUIDStr)

	// 4. 获取全局连接管理器并寻找接收者连接
	connManager := global.GlobalServer.GetConnManager()
	fmt.Printf("[系统状态] 当前在线连接数: %d\n", connManager.Size())

	// 5. 遍历所有连接，查找目标用户
	foundOnline := false
	for _, conn := range connManager.All() {
		// 获取连接关联的用户ID
		if userIDProp, err := conn.GetProperty("userID"); err == nil {
			// 确保类型一致进行比较
			userIDUint, ok := userIDProp.(uint)
			if ok && userIDUint == toUserIDUint { // Compare uint with uint
				fmt.Printf("[消息投递] 用户 %s (ID: %d) 在线，直接发送消息\n", toUsernameStr, toUserIDUint)
				// 目标用户在线，直接转发消息
				err := conn.SendMsg(protocol.MsgIDTextMsg, msgData)
				if err != nil {
					fmt.Printf("[消息发送] 发送失败: %v, 将保存为离线消息\n", err)
				} else {
					foundOnline = true
					break
				}
			}
		}
	}

	if !foundOnline {
		fmt.Printf("[离线存储] 用户 %s (ID: %d) 不在线或消息发送失败，存储为离线消息\n", toUsernameStr, toUserIDUint)
		// 3. 只有当用户不在线或消息发送失败时，才保存到Redis（历史记录和离线消息）
		err = global.MessageService.SaveMessage(fromUserIDUint, toUserIDUint, msgData) // Pass uint IDs
		if err != nil {
			fmt.Printf("[Redis消息] 保存消息失败: %v\n", err)
			// TODO: Handle this error, maybe log or retry
			return
		}
	} else {
		// 用户在线并且消息发送成功，只保存历史记录不保存离线消息
		err = global.MessageService.SaveHistoryOnly(fromUserIDUint, toUserIDUint, msgData) // Pass uint IDs
		if err != nil {
			fmt.Printf("[Redis消息] 保存历史记录失败: %v\n", err)
			// TODO: Handle this error
			return
		}
		fmt.Printf("[历史记录] 用户 %s (ID: %d) 在线且消息发送成功，只记录历史不保存离线消息\n", toUsernameStr, toUserIDUint)
	}
}
