package router

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Xaytick/chat-zinx/chat-server/global"
	// "github.com/Xaytick/chat-zinx/chat-server/pkg/middleware" // Token生成移至Service层或按需处理
	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/service" // 用于错误比较
	"github.com/Xaytick/zinx/ziface"
	"github.com/Xaytick/zinx/znet"
)

// RegisterRouter 处理注册请求
type RegisterRouter struct {
	znet.BaseRouter
	// auth *middleware.AuthMiddleware // 移除非必要依赖
}

// NewRegisterRouter 创建新的注册路由 (如果不需要内部状态如auth, 可以移除)
// func NewRegisterRouter() *RegisterRouter {
// 	return &RegisterRouter{}
// }

func (r *RegisterRouter) Handle(request ziface.IRequest) {
	var registerReq model.UserRegisterReq
	if err := json.Unmarshal(request.GetData(), &registerReq); err != nil {
		sendRegisterResponse(request, 1, "请求数据格式错误", nil)
		return
	}

	if registerReq.Username == "" || registerReq.Password == "" {
		sendRegisterResponse(request, 2, "用户名和密码不能为空", nil)
		return
	}
	// Email 验证可以做得更细致，例如使用正则表达式
	if registerReq.Email == "" { // 简单检查，model中已有 binding:"required,email"
		sendRegisterResponse(request, 2, "邮箱不能为空", nil)
		return
	}

	user, err := global.UserService.Register(&registerReq)
	if err != nil {
		errMsg := "注册失败: " + err.Error()
		var code uint32 = 4
		// 检查是否是 service 层定义的特定错误，例如用户已存在
		// 假设 service.Register 返回的 error 会包含 "username already exists" 或类似的具体信息
		// 或者 service 层可以定义并返回如 service.ErrUserAlreadyExists
		if errors.Is(err, service.ErrUserAlreadyExists) { // 确保这个错误在 service 层被定义和返回
			errMsg = "用户已存在"
			code = 3
		}
		sendRegisterResponse(request, code, errMsg, nil)
		return
	}

	// 注册成功
	// 业务决定注册后是否自动登录并返回Token。
	// 当前 UserRegisterResponse 的 Token 字段是 omitempty。
	// 如果需要注册后自动登录, 则在此处调用 Login 服务获取Token。
	// tokenString, _, err := global.UserService.Login(&model.UserLoginReq{Username: user.Username, Password: registerReq.Password})
	// if err != nil { ... handle error ... }

	responseData := model.UserRegisterResponse{
		ID:       user.ID,
		UserUUID: user.UserUUID,
		Username: user.Username,
		Email:    user.Email,
		Avatar:   user.Avatar,
		// Token:    tokenString, // 如果上面获取了token
	}

	sendRegisterResponse(request, 0, "注册成功", responseData)
	fmt.Printf("User %s registered successfully, ID: %d, UUID: %s\n", user.Username, user.ID, user.UserUUID)
}

func sendRegisterResponse(request ziface.IRequest, code uint32, message string, data interface{}) {
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
	request.GetConnection().SendMsg(protocol.MsgIDRegisterResp, jsonData)
}
