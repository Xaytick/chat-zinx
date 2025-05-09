package router

import (
	"encoding/json"
	"fmt"

	"github.com/Xaytick/chat-zinx/chat-server/global"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/service"
	"github.com/Xaytick/zinx/ziface"
	"github.com/Xaytick/zinx/znet"
)

// RegisterRouter 处理注册请求
type RegisterRouter struct {
	znet.BaseRouter
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
			sendRegisterResponse(request, 4, "注册失败: " + err.Error(), nil)
		}
		return
	}

	// 构造返回数据(不含敏感信息)
	responseData := &model.UserRegisterResponse{
		UserID:    user.UserID,
		Username:  user.Username,
		Email:     user.Email,
	}

	// 4. 返回成功响应
	sendRegisterResponse(request, 0, "注册成功", responseData)
	fmt.Printf("用户 %s 注册成功, ID: %s\n", user.Username, user.UserID)

	
}

func sendRegisterResponse(requst ziface.IRequest, code uint32, message string, data interface{}) {
	response := map[string]interface{}{
		"code": code,
		"msg": message,
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