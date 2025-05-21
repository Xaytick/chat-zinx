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

// HistoryMsgRouter 处理历史消息请求
type HistoryMsgRouter struct {
	znet.BaseRouter
}

func (r *HistoryMsgRouter) Handle(request ziface.IRequest) {
	// 1. 获取当前用户ID
	fromUserIDProp, err := request.GetConnection().GetProperty("userID")
	if err != nil || fromUserIDProp == nil {
		sendHistoryResponse(request, 1, "用户未登录或会话无效", nil)
		return
	}
	fromUserIDUint, ok := fromUserIDProp.(uint)
	if !ok {
		fmt.Println("[历史消息] fromUserID 类型错误 on connection property")
		sendHistoryResponse(request, 1, "内部错误：用户ID无效", nil)
		return
	}

	// 2. 解析请求
	var req model.LegacyHistoryMsgReq
	if err := json.Unmarshal(request.GetData(), &req); err != nil {
		sendHistoryResponse(request, 2, "请求格式错误", nil)
		return
	}

	// 3. 确定目标用户
	var targetUser *model.User
	if req.TargetUserUUID != "" {
		targetUser, err = global.UserService.GetUserByUUID(req.TargetUserUUID)
		if err != nil {
			if errors.Is(err, service.ErrUserNotFound) {
				sendHistoryResponse(request, 3, fmt.Sprintf("目标用户UUID %s 不存在", req.TargetUserUUID), nil)
			} else {
				fmt.Printf("[历史消息] 根据UUID查找目标用户 %s 失败: %v\n", req.TargetUserUUID, err)
				sendHistoryResponse(request, 3, "查找目标用户失败", nil)
			}
			return
		}
	} else if req.TargetUsername != "" {
		targetUser, err = global.UserService.GetUserByUsername(req.TargetUsername)
		if err != nil {
			if errors.Is(err, service.ErrUserNotFound) {
				sendHistoryResponse(request, 3, fmt.Sprintf("目标用户 %s 不存在", req.TargetUsername), nil)
			} else {
				fmt.Printf("[历史消息] 根据Username查找目标用户 %s 失败: %v\n", req.TargetUsername, err)
				sendHistoryResponse(request, 3, "查找目标用户失败", nil)
			}
			return
		}
	} else {
		sendHistoryResponse(request, 3, "必须提供目标用户的UUID或用户名", nil)
		return
	}

	if targetUser == nil {
		sendHistoryResponse(request, 3, "无法确定目标用户", nil)
		return
	}
	targetUserIDUint := targetUser.ID

	// 4. 设置默认限制
	if req.Limit <= 0 {
		req.Limit = 50
	} else if req.Limit > 200 {
		req.Limit = 200
	}

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
	response := model.LegacyHistoryMsgResp{
		Code:    code,
		Message: message,
		Data:    data,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("序列化历史消息响应失败: %v\n", err)
		return
	}
	request.GetConnection().SendMsg(protocol.MsgIDHistoryMsgResp, jsonData)
}
