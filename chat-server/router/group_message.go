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

	// 3. 构建推送消息
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

	// 4. 向群内其他在线成员推送消息
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

	// 5. (可选) 保存群消息到持久化存储 - 暂时跳过
	// if err := global.MessageService.SaveGroupMessage(reqPayload.GroupID, userID, pushData); err != nil {
	// 	 fmt.Printf("[GroupMsgRouter] UserID %d: Failed to save group message for GroupID %d: %v\n", userID, reqPayload.GroupID, err)
	// 	 // Non-critical error for now, client already got the push (if online)
	// }

	// 6. 向发送者回复成功
	successResp := model.GroupTextMsgResp{Status: 0, MsgID: "some-unique-msg-id"} // TODO: Generate actual unique msg ID if needed
	successRespData, _ := json.Marshal(successResp)
	conn.SendMsg(protocol.MsgIDGroupTextMsgResp, successRespData)
}
