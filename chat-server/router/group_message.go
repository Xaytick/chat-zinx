package router

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Xaytick/chat-zinx/chat-server/global"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
	"github.com/Xaytick/zinx/ziface"
	"github.com/Xaytick/zinx/znet"
)

// GroupTextMsgRouter 处理群组文本消息的路由
type GroupTextMsgRouter struct {
	znet.BaseRouter
}

// Handle 处理客户端发送的群组文本消息请求
func (r *GroupTextMsgRouter) Handle(request ziface.IRequest) {
	conn := request.GetConnection()

	userIDVal, err := conn.GetProperty("userID")
	if err != nil {
		fmt.Println("[GroupMsgRouter] Failed to get userID from connection properties:", err)
		// TODO: Send error response to client if possible, or just close connection
		return
	}
	userID := userIDVal.(uint)

	userUUIDVal, _ := conn.GetProperty("userUUID") // Assuming these exist if userID exists
	userUUID := userUUIDVal.(string)
	usernameVal, _ := conn.GetProperty("username")
	username := usernameVal.(string)

	var reqPayload model.GroupTextMsgReq
	if err := json.Unmarshal(request.GetData(), &reqPayload); err != nil {
		fmt.Printf("[GroupMsgRouter] UserID %d: Failed to unmarshal GroupTextMsgReq: %v\n", userID, err)
		// Send error response
		resp := model.GroupTextMsgResp{Status: 1, Error: "Invalid request format"}
		respData, _ := json.Marshal(resp)
		conn.SendMsg(protocol.MsgIDGroupTextMsgResp, respData)
		return
	}

	fmt.Printf("[GroupMsgRouter] UserID %d (%s) sending message to GroupID %d: %s\n", userID, username, reqPayload.GroupID, reqPayload.Content)

	// 1. 验证用户是否为群组成员
	isMember, err := global.GroupService.IsUserInGroup(userID, uint(reqPayload.GroupID))
	if err != nil {
		fmt.Printf("[GroupMsgRouter] UserID %d: Error checking group membership for GroupID %d: %v\n", userID, reqPayload.GroupID, err)
		resp := model.GroupTextMsgResp{Status: 2, Error: "Failed to verify group membership"}
		respData, _ := json.Marshal(resp)
		conn.SendMsg(protocol.MsgIDGroupTextMsgResp, respData)
		return
	}
	if !isMember {
		fmt.Printf("[GroupMsgRouter] UserID %d is not a member of GroupID %d. Message rejected.\n", userID, reqPayload.GroupID)
		resp := model.GroupTextMsgResp{Status: 3, Error: "You are not a member of this group"}
		respData, _ := json.Marshal(resp)
		conn.SendMsg(protocol.MsgIDGroupTextMsgResp, respData)
		return
	}

	// 2. 获取群组成员ID列表
	memberIDs, err := global.GroupService.GetGroupMemberIDs(uint(reqPayload.GroupID))
	if err != nil {
		fmt.Printf("[GroupMsgRouter] UserID %d: Failed to get member IDs for GroupID %d: %v\n", userID, reqPayload.GroupID, err)
		resp := model.GroupTextMsgResp{Status: 4, Error: "Failed to retrieve group members"}
		respData, _ := json.Marshal(resp)
		conn.SendMsg(protocol.MsgIDGroupTextMsgResp, respData)
		return
	}

	// 3. 保存消息到数据库
	msgID, err := global.MessageService.SaveGroupMessage(
		uint(reqPayload.GroupID),
		userID,
		userUUID,
		username,
		reqPayload.Content,
		"text", // 默认为文本消息类型
	)
	if err != nil {
		fmt.Printf("[GroupMsgRouter] UserID %d: Failed to save message to database for GroupID %d: %v\n", userID, reqPayload.GroupID, err)
		// 这是一个非关键错误，我们仍然可以继续处理，但需要记录日志
	}

	// 4. 构建推送消息
	pushMsg := model.GroupTextMsgPush{
		GroupID:      reqPayload.GroupID,
		FromUserID:   userID,
		FromUserUUID: userUUID,
		FromUsername: username,
		Content:      reqPayload.Content,
		Timestamp:    time.Now().Unix(),
	}
	pushData, err := json.Marshal(pushMsg)
	if err != nil {
		fmt.Printf("[GroupMsgRouter] UserID %d: Failed to marshal GroupTextMsgPush for GroupID %d: %v\n", userID, reqPayload.GroupID, err)
		// This is an internal server error, might not need to send specific error to client here,
		// but a general success ack (step 4) might fail or be misleading.
		// For now, let's send a generic error back if marshalling fails.
		resp := model.GroupTextMsgResp{Status: 5, Error: "Internal server error preparing message"}
		respData, _ := json.Marshal(resp)
		conn.SendMsg(protocol.MsgIDGroupTextMsgResp, respData)
		return
	}

	// 5. 向群内其他在线成员推送消息
	connMgr := global.GlobalServer.GetConnManager() // Attempt to call the GetConnManager method on *znet.Server
	membersNotified := 0
	for _, memberID := range memberIDs {
		if memberID == userID { // 不给自己推送
			continue
		}
		targetConn := connMgr.GetConnByUserID(memberID) // Assuming GetConnByUserID is implemented
		if targetConn != nil {
			fmt.Printf("[GroupMsgRouter] Pushing message from UserID %d to UserID %d in GroupID %d\n", userID, memberID, reqPayload.GroupID)
			err := targetConn.SendMsg(protocol.MsgIDGroupTextMsgPush, pushData)
			if err != nil {
				fmt.Printf("[GroupMsgRouter] Failed to send group message to UserID %d in GroupID %d: %v\n", memberID, reqPayload.GroupID, err)
				// TODO: Handle case where sending to a member fails (e.g. connection closed suddenly)
			} else {
				membersNotified++
			}
		} else {
			// fmt.Printf("[GroupMsgRouter] UserID %d in GroupID %d is offline or not found in ConnMgr.\n", memberID, reqPayload.GroupID)
		}
	}
	fmt.Printf("[GroupMsgRouter] Message from UserID %d to GroupID %d pushed to %d online members.\n", userID, reqPayload.GroupID, membersNotified)

	// 6. 向发送者回复成功
	successResp := model.GroupTextMsgResp{Status: 0, MsgID: msgID}
	successRespData, _ := json.Marshal(successResp)
	conn.SendMsg(protocol.MsgIDGroupTextMsgResp, successRespData)
}

// GroupHistoryMsgRouter 处理获取群组历史消息的路由
type GroupHistoryMsgRouter struct {
	znet.BaseRouter
}

// Handle 处理获取群组历史消息请求
func (r *GroupHistoryMsgRouter) Handle(request ziface.IRequest) {
	conn := request.GetConnection()

	userIDVal, err := conn.GetProperty("userID")
	if err != nil {
		fmt.Println("[GroupHistoryMsgRouter] Failed to get userID from connection properties:", err)
		errorResp, _ := json.Marshal(map[string]string{"error": "用户未登录"})
		conn.SendMsg(protocol.MsgIDGroupHistoryMsgResp, errorResp)
		return
	}
	userID := userIDVal.(uint)

	var req model.GroupHistoryMsgReq
	if err := json.Unmarshal(request.GetData(), &req); err != nil {
		fmt.Printf("[GroupHistoryMsgRouter] UserID %d: Failed to unmarshal request: %v\n", userID, err)
		errorResp, _ := json.Marshal(map[string]string{"error": "请求格式错误"})
		conn.SendMsg(protocol.MsgIDGroupHistoryMsgResp, errorResp)
		return
	}

	// 默认值处理
	if req.Limit <= 0 {
		req.Limit = 20 // 默认获取20条消息
	} else if req.Limit > 100 {
		req.Limit = 100 // 最多获取100条消息
	}

	// 获取历史消息
	resp, err := global.MessageService.GetGroupHistory(userID, req.GroupID, req.LastID, req.Limit)
	if err != nil {
		fmt.Printf("[GroupHistoryMsgRouter] UserID %d: Failed to get history for GroupID %d: %v\n", userID, req.GroupID, err)
		errorResp, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("获取历史消息失败：%s", err.Error())})
		conn.SendMsg(protocol.MsgIDGroupHistoryMsgResp, errorResp)
		return
	}

	// 发送响应
	respData, _ := json.Marshal(resp)
	conn.SendMsg(protocol.MsgIDGroupHistoryMsgResp, respData)
	fmt.Printf("[GroupHistoryMsgRouter] Retrieved %d messages for GroupID %d (UserID %d)\n", len(resp.Messages), req.GroupID, userID)
}
