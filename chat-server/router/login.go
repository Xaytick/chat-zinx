package router

import (
	"encoding/json"

	"github.com/Xaytick/chat-server/pkg/protocol"
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
	userID := payload.Username
	// 验证成功后，将 userID 写入 Conn 属性
	request.GetConnection().SetProperty("userID", userID)
	// 并回复客户端登录结果
	resp := map[string]interface{}{"code": 0, "msg": "登录成功"}
	data, _ := json.Marshal(resp)
	request.GetConnection().SendMsg(protocol.MsgIDLoginReq, data)
}

// 登录后处理
func (lr *LoginRouter) PostHandle(request ziface.IRequest) {
	// 这里可以记录登录日志、踢下线等
}
