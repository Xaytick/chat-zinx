package router

import (
	"encoding/json"
	"fmt"

	"github.com/Xaytick/chat-zinx/chat-server/global"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
	"github.com/Xaytick/zinx/ziface"
	"github.com/Xaytick/zinx/znet"
)

// 历史消息请求结构
type HistoryMsgReq struct {
	TargetUserUUID string `json:"target_user_uuid"`
	Limit          int    `json:"limit"`
}

// 历史消息响应结构
type HistoryMsgResp struct {
	Code    uint32                   `json:"code"`
	Message string                   `json:"message"`
	Data    []map[string]interface{} `json:"data"`
}

// HistoryMsgRouter 处理历史消息请求
type HistoryMsgRouter struct {
	znet.BaseRouter
}

func (r *HistoryMsgRouter) Handle(request ziface.IRequest) {
	// 1. 获取当前用户ID
	fromUserIDProp, err := request.GetConnection().GetProperty("userID")
	if err != nil || fromUserIDProp == nil {
		sendHistoryResponse(request, 1, "未登录用户", nil)
		return
	}
	fromUserIDUint, ok := fromUserIDProp.(uint)
	if !ok {
		fmt.Println("[历史消息] fromUserID 类型错误 on connection property")
		sendHistoryResponse(request, 1, "内部错误：用户ID无效", nil)
		return
	}

	// 2. 解析请求
	var req HistoryMsgReq
	if err := json.Unmarshal(request.GetData(), &req); err != nil {
		sendHistoryResponse(request, 2, "请求格式错误", nil)
		return
	}

	// 3. 验证参数
	if req.TargetUserUUID == "" {
		sendHistoryResponse(request, 3, "目标用户UUID不能为空", nil)
		return
	}

	// 设置默认限制
	if req.Limit <= 0 {
		req.Limit = 50
	} else if req.Limit > 200 {
		req.Limit = 200
	}

	// 4. 根据 TargetUserUUID 查找目标用户以获取其 uint ID
	targetUser, err := global.UserService.GetUserByUUID(req.TargetUserUUID)
	if err != nil || targetUser == nil {
		fmt.Printf("[历史消息] 查找目标用户 %s 失败: %v\n", req.TargetUserUUID, err)
		// Check if it might be a username or numeric ID if GetUserByUUID fails
		// For now, assume UUID is provided or fail.
		sendHistoryResponse(request, 3, "目标用户不存在", nil)
		return
	}
	targetUserIDUint := targetUser.ID

	// 5. 获取历史消息
	messages, err := global.MessageService.GetHistoryMessages(fromUserIDUint, targetUserIDUint, req.Limit)
	if err != nil {
		fmt.Printf("[历史消息] 获取失败: %v\n", err)
		sendHistoryResponse(request, 4, "获取历史消息失败", nil)
		return
	}

	// 6. 返回历史消息
	sendHistoryResponse(request, 0, "成功", messages)
}

// 发送历史消息响应
func sendHistoryResponse(request ziface.IRequest, code uint32, message string, data []map[string]interface{}) {
	response := HistoryMsgResp{
		Code:    code,
		Message: message,
		Data:    data,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("序列化失败: %v\n", err)
		return
	}
	request.GetConnection().SendMsg(protocol.MsgIDHistoryMsgResp, jsonData)
}
