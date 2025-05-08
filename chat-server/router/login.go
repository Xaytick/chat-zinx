package router

import (
	"encoding/json"
	"fmt"

	"github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/storage"
	"github.com/Xaytick/zinx/ziface"
	"github.com/Xaytick/zinx/znet"
)

// 登录路由, 处理登录请求
type LoginRouter struct {
	znet.BaseRouter
}

// 登录前验证
func (r *LoginRouter) PreHandle(request ziface.IRequest) {
	// 这里可以验证token签名、IP黑名单等
}

// 核心登录业务
func (lr *LoginRouter) Handle(request ziface.IRequest) {
	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	// 解析请求数据
	if err := json.Unmarshal(request.GetData(), &payload); err != nil {
		request.GetConnection().SendMsg(protocol.MsgIDLoginReq, []byte("解析请求数据失败"))
		return
	}
	// TODO: 调用用户服务验证用户名/密码
	// 验证成功后，将 userID 写入 Conn 属性
	userID := payload.Username
	request.GetConnection().SetProperty("userID", userID)

	// 1. 先回复客户端登录结果
	resp := map[string]interface{}{"code": 0, "msg": "登录成功"}
	data, _ := json.Marshal(resp)
	request.GetConnection().SendMsg(protocol.MsgIDLoginReq, data)

	// 2. 如果用户有离线消息，再推送离线消息
	if storage.HasOfflineMessages(userID) {
		offlineMsgs := storage.GetOfflineMessages(userID)
		fmt.Println("用户", userID, "有离线消息，推送", len(offlineMsgs), "条")
		for _, msgContent := range offlineMsgs {
			request.GetConnection().SendMsg(protocol.MsgIDTextMsg, []byte(msgContent))
		}
	}
}

// 登录后处理
func (lr *LoginRouter) PostHandle(request ziface.IRequest) {
	// 这里可以记录登录日志、踢下线等
}
