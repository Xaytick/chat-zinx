package router

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Xaytick/chat-zinx/chat-server/global"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/service"
	"github.com/Xaytick/zinx/ziface"
	"github.com/Xaytick/zinx/znet"
)

// LoginRouter 处理登录请求
type LoginRouter struct {
	znet.BaseRouter
}

// NewLoginRouter 创建新的登录路由
func NewLoginRouter() *LoginRouter {
	return &LoginRouter{}
}

// 登录前验证
func (r *LoginRouter) PreHandle(request ziface.IRequest) {
	// 这里可以验证token签名、IP黑名单等
}

// 核心登录业务
func (lr *LoginRouter) Handle(request ziface.IRequest) {
	var loginReq model.UserLoginReq
	if err := json.Unmarshal(request.GetData(), &loginReq); err != nil {
		sendLoginResponse(request, 1, "请求数据格式错误", nil)
		return
	}
	if loginReq.Username == "" || loginReq.Password == "" {
		sendLoginResponse(request, 2, "用户名或密码不能为空", nil)
		return
	}

	// 调用用户服务验证用户名/密码
	tokenString, user, err := global.UserService.Login(&loginReq)

	if err != nil {
		// 尝试自动注册的逻辑 (如果仍然需要)
		// 注意：如果Login失败是因为用户不存在，且希望自动注册，这里的错误处理需要调整
		// 当前 UserService.Login 遇到用户不存在直接返回错误，不包含自动注册逻辑
		// 如果需要自动注册，应该在 Login 失败后，显式调用 Register
		fmt.Printf("Login failed for %s: %v\n", loginReq.Username, err)
		// 根据错误类型发送不同的消息
		errMsg := "登录失败"
		var code uint32 = 5
		if errors.Is(err, service.ErrUserNotFound) { // 假设 service 层定义了 ErrUserNotFound
			code = 3
			errMsg = "用户不存在"
		} else if errors.Is(err, service.ErrPasswordIncorrect) { // 假设 service 层定义了 ErrPasswordIncorrect
			code = 4
			errMsg = "密码错误"
		}
		sendLoginResponse(request, code, errMsg, nil)
		return
	}

	// 登录成功
	request.GetConnection().SetProperty("userID", user.ID) // 使用 uint 类型的 ID
	request.GetConnection().SetProperty("userUUID", user.UserUUID)
	request.GetConnection().SetProperty("username", user.Username)

	fmt.Printf("User %s (ID: %d, UUID: %s) logged in successfully.\n", user.Username, user.ID, user.UserUUID)

	connMgr := global.GlobalServer.GetConnManager()
	connMgr.SetConnByUserID(request.GetConnection().GetConnID(), user.ID)

	// 构造返回数据
	responseData := model.UserLoginResponse{
		ID:        user.ID,
		UserUUID:  user.UserUUID,
		Username:  user.Username,
		Email:     user.Email,
		Avatar:    user.Avatar,
		LastLogin: user.LastLogin, // 已是 time.Time 类型
		Token:     tokenString,
	}

	sendLoginResponse(request, 0, "登录成功", responseData)

	// --- 离线消息处理逻辑 (保持不变或根据 userID (uint) 调整) ---
	userIDForOffline := user.ID // 使用 uint 类型的 ID 获取离线消息
	processedMsgs := make(map[string]bool)
	offlineMsgs := [][]byte{}

	// 1. 先获取用户ID的离线消息
	if global.MessageService.HasOfflineMessages(userIDForOffline) { // 直接传递 uint 类型的 userIDForOffline
		msgs, err := global.MessageService.GetOfflineMessages(userIDForOffline) // 直接传递 uint 类型的 userIDForOffline
		if err != nil {
			fmt.Printf("[Redis错误] 获取离线消息失败 for ID %d: %v\n", userIDForOffline, err)
		} else {
			for _, msg := range msgs {
				msgKey := fmt.Sprintf("%x", msg)
				if !processedMsgs[msgKey] {
					processedMsgs[msgKey] = true
					offlineMsgs = append(offlineMsgs, msg)
				}
			}
		}
	}

	// 2. 用户名下的离线消息 (如果仍然需要，通常在用户能用ID识别后，应迁移到ID下)
	// if global.MessageService.HasOfflineMessages(user.Username) { ... }

	// 3. 如果有离线消息，推送给用户
	if len(offlineMsgs) > 0 {
		fmt.Printf("[离线消息] 用户 %s(ID:%d) 共有 %d 条离线消息待推送\n",
			user.Username, user.ID, len(offlineMsgs))

		for _, msgData := range offlineMsgs {
			request.GetConnection().SendMsg(protocol.MsgIDTextMsg, msgData)
		}
	}
}

// 登录后处理
func (lr *LoginRouter) PostHandle(request ziface.IRequest) {
	// 这里可以记录登录日志、踢下线等
}

// 发送登录响应
func sendLoginResponse(request ziface.IRequest, code uint32, msg string, data interface{}) {
	response := map[string]interface{}{
		"code": code,
		"msg":  msg,
	}
	if data != nil {
		response["data"] = data
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("序列化失败: %v\n", err)
		return
	}
	request.GetConnection().SendMsg(protocol.MsgIDLoginResp, jsonData)
}
