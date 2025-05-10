package router

import (
	"encoding/json"
	"fmt"

	"github.com/Xaytick/chat-zinx/chat-server/global"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
	"github.com/Xaytick/zinx/ziface"
	"github.com/Xaytick/zinx/znet"
)

// 聊天关系响应结构
type ChatRelationResp struct {
	Code    uint32   `json:"code"`
	Message string   `json:"message"`
	Data    []string `json:"data"` // 用户ID列表
}

// ChatRelationRouter 处理聊天关系请求
type ChatRelationRouter struct {
	znet.BaseRouter
}

func (r *ChatRelationRouter) Handle(request ziface.IRequest) {
	// 1. 获取当前用户ID
	userID, err := request.GetConnection().GetProperty("userID")
	if err != nil {
		sendChatRelationResponse(request, 1, "未登录用户", nil)
		return
	}

	// 2. 获取用户的聊天关系列表
	relations, err := global.MessageService.GetChatRelations(userID.(string))
	if err != nil {
		fmt.Printf("[聊天关系] 获取失败: %v\n", err)
		sendChatRelationResponse(request, 2, "获取聊天关系失败", nil)
		return
	}

	// 3. 返回聊天关系列表
	sendChatRelationResponse(request, 0, "成功", relations)
}

// 发送聊天关系响应
func sendChatRelationResponse(request ziface.IRequest, code uint32, message string, data []string) {
	response := ChatRelationResp{
		Code:    code,
		Message: message,
		Data:    data,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("序列化失败: %v\n", err)
		return
	}
	request.GetConnection().SendMsg(protocol.MsgIDChatRelationResp, jsonData)
}
