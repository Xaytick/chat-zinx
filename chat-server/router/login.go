package router

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Xaytick/chat-zinx/chat-server/global"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/middleware"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/service"
	"github.com/Xaytick/zinx/ziface"
	"github.com/Xaytick/zinx/znet"
)

// LoginRouter 处理登录请求
type LoginRouter struct {
	znet.BaseRouter
	auth *middleware.AuthMiddleware
}

// NewLoginRouter 创建新的登录路由
func NewLoginRouter() *LoginRouter {
	return &LoginRouter{
		auth: middleware.NewAuthMiddleware(),
	}
}

// 登录前验证
func (r *LoginRouter) PreHandle(request ziface.IRequest) {
	// 这里可以验证token签名、IP黑名单等
}

// 核心登录业务
func (lr *LoginRouter) Handle(request ziface.IRequest) {
	// 解析请求数据
	var loginReq model.UserLoginReq
	if err := json.Unmarshal(request.GetData(), &loginReq); err != nil {
		sendLoginResponse(request, 1, "请求数据格式错误", nil)
		return
	}
	// 参数验证
	if loginReq.Username == "" || loginReq.Password == "" {
		sendLoginResponse(request, 2, "用户名或密码不能为空", nil)
		return
	}

	// 调用用户服务验证用户名/密码
	user, err := global.UserService.Login(loginReq)
	if err != nil {
		var code uint32
		var msg string
		switch err {
		// 用户不存在
		case service.ErrUserNotFound:
			// 尝试自动注册
			if registerUser, registerErr := global.UserService.Register(model.UserRegisterReq{
				Username: loginReq.Username,
				Password: loginReq.Password,
			}); registerErr == nil {
				// 自动注册成功
				user = registerUser
				fmt.Printf("用户 %s 自动注册成功，ID: %s\n", user.Username, user.UserID)
				break
			} else {
				code = 3
				msg = "用户不存在，且自动注册失败"
			}
		case service.ErrPasswordIncorrect:
			code = 4
			msg = "密码错误"
		default:
			code = 5
			msg = "登录失败"
		}
		if user == nil {
			sendLoginResponse(request, code, msg, nil)
			return
		}
	}
	// 登录成功, 绑定用户ID到连接
	userID := user.UserID
	request.GetConnection().SetProperty("userID", userID)
	request.GetConnection().SetProperty("username", user.Username)

	// 确保auth中间件已初始化
	if lr.auth == nil {
		lr.auth = middleware.NewAuthMiddleware()
	}

	// 生成JWT令牌
	token, err := lr.auth.GenerateToken(userID, user.Username)
	if err != nil {
		fmt.Printf("[登录] 生成Token失败: %v\n", err)
		sendLoginResponse(request, 6, "生成令牌失败", nil)
		return
	}

	// 如果启用了Redis会话验证，保存会话
	lr.auth.SaveSession(userID, token)

	// 构造返回数据
	responseData := &model.UserLoginResponse{
		UserID:    userID,
		Username:  user.Username,
		Email:     user.Email,
		LastLogin: user.LastLogin.Format(time.DateTime),
		Token:     token, // 添加令牌
	}

	// 先回复客户端登录结果
	sendLoginResponse(request, 0, "登录成功", responseData)

	// 如果用户有离线消息，再推送离线消息
	// 使用map记录已处理的消息，避免重复推送
	processedMsgs := make(map[string]bool)
	offlineMsgs := [][]byte{}

	// 1. 先获取用户ID的离线消息（优先级更高）
	if global.MessageService.HasOfflineMessages(userID) {
		msgs, err := global.MessageService.GetOfflineMessages(userID)
		if err != nil {
			fmt.Printf("[Redis错误] 获取离线消息失败: %v\n", err)
		} else {
			for _, msg := range msgs {
				// 使用消息内容的哈希值作为唯一标识，避免重复
				msgKey := fmt.Sprintf("%x", msg)
				if !processedMsgs[msgKey] {
					processedMsgs[msgKey] = true
					offlineMsgs = append(offlineMsgs, msg)
				}
			}
		}
	}

	// 2. 再检查用户名下是否有离线消息（可能是在用户注册前发送的）
	if global.MessageService.HasOfflineMessages(user.Username) {
		fmt.Printf("[离线消息] 发现以用户名 %s 保存的离线消息\n", user.Username)
		msgs, err := global.MessageService.GetOfflineMessages(user.Username)
		if err != nil {
			fmt.Printf("[Redis错误] 获取离线消息失败: %v\n", err)
		} else {
			for _, msg := range msgs {
				// 使用消息内容的哈希值作为唯一标识，避免重复
				msgKey := fmt.Sprintf("%x", msg)
				if !processedMsgs[msgKey] {
					processedMsgs[msgKey] = true
					offlineMsgs = append(offlineMsgs, msg)
				}
			}
		}
	}

	// 3. 如果有离线消息，推送给用户
	if len(offlineMsgs) > 0 {
		fmt.Printf("[离线消息] 用户 %s(ID:%s) 共有 %d 条离线消息待推送\n",
			user.Username, userID, len(offlineMsgs))

		for _, msgData := range offlineMsgs {
			// 直接推送消息，消息中已包含发送者ID
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
