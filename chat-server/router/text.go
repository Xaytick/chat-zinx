package router

import (
	"encoding/json"
	"fmt"

	"github.com/Xaytick/chat-zinx/chat-server/global"
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
	var msg protocol.TextMsg
	if err := json.Unmarshal(request.GetData(), &msg); err != nil {
		// 解析失败，返回错误
		fmt.Println("消息解析失败", err)
		return
	}

	// 2. 获取全局连接管理器
	connManager := global.GlobalServer.GetConnManager()

	// 3. 遍历所有连接，查找目标用户
	found := false
	for _, conn := range connManager.All() {
		if v, err := conn.GetProperty("userID"); err == nil && v == msg.ToUserID {
			// 4. 目标用户在线，直接转发消息
			conn.SendMsg(protocol.MsgIDTextMsg, request.GetData())
			found = true
			break
		}
	}

	if !found {
		// 5. 目标用户不在线，存储为离线消息
		storage.SaveOfflineMsg(msg.ToUserID, string(request.GetData()))
	}
}
