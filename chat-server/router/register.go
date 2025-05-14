package router

import (
	"encoding/json"
	"fmt"

	"github.com/Xaytick/chat-zinx/chat-server/global"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/middleware"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/service"
	"github.com/Xaytick/zinx/ziface"
	"github.com/Xaytick/zinx/znet"
)

// RegisterRouter 处理注册请求
type RegisterRouter struct {
	znet.BaseRouter
	auth *middleware.AuthMiddleware
}

// NewRegisterRouter 创建新的注册路由
func NewRegisterRouter() *RegisterRouter {
	return &RegisterRouter{
		auth: middleware.NewAuthMiddleware(),
	}
}

func (r *RegisterRouter) Handle(request ziface.IRequest) {
	// 1. 解析消息体
	var registerReq model.UserRegisterReq
	if err := json.Unmarshal(request.GetData(), &registerReq); err != nil {
		// 解析失败，返回错误
		sendRegisterResponse(request, 1, "请求数据格式错误", nil)
		return
	}

	// 2. 参数验证
	if registerReq.Username == "" || registerReq.Password == "" {
		sendRegisterResponse(request, 2, "用户名和密码不能为空", nil)
		return
	}

	// 3. 调用用户服务注册
	user, err := global.UserService.Register(registerReq)
	if err != nil {
		if err == service.ErrUserAlreadyExists {
			// 用户已存在，返回错误
			sendRegisterResponse(request, 3, "用户已存在", nil)
		} else {
			// 其他错误，返回错误
			sendRegisterResponse(request, 4, "注册失败: "+err.Error(), nil)
		}
		return
	}

	// 确保auth中间件已初始化
	if r.auth == nil {
		r.auth = middleware.NewAuthMiddleware()
	}

	// 生成JWT令牌
	token, err := r.auth.GenerateToken(user.UserID, user.Username)
	if err != nil {
		fmt.Printf("[注册] 生成Token失败: %v\n", err)
		sendRegisterResponse(request, 5, "生成令牌失败", nil)
		return
	}

	// 如果启用了Redis会话验证，保存会话
	r.auth.SaveSession(user.UserID, token)

	// 构造返回数据(不含敏感信息)
	responseData := &model.UserRegisterResponse{
		UserID:   user.UserID,
		Username: user.Username,
		Email:    user.Email,
		Token:    token,
	}

	// 4. 返回成功响应
	sendRegisterResponse(request, 0, "注册成功", responseData)
	fmt.Printf("用户 %s 注册成功, ID: %s\n", user.Username, user.UserID)
}

func sendRegisterResponse(requst ziface.IRequest, code uint32, message string, data interface{}) {
	response := map[string]interface{}{
		"code": code,
		"msg":  message,
	}
	if data != nil {
		response["data"] = data
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("序列化失败: %v\n", err)
		return
	}
	requst.GetConnection().SendMsg(protocol.MsgIDRegisterResp, jsonData)
}
