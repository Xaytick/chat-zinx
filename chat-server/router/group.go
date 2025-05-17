package router

import (
	"encoding/json"
	"fmt"

	"github.com/Xaytick/chat-zinx/chat-server/global"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
	"github.com/Xaytick/zinx/ziface"
	"github.com/Xaytick/zinx/znet"
)

// --- CreateGroupRouter --- //
type CreateGroupRouter struct {
	znet.BaseRouter
}

func (r *CreateGroupRouter) Handle(request ziface.IRequest) {
	userID, err := request.GetConnection().GetProperty("userID")
	if err != nil || userID == nil {
		fmt.Println("CreateGroupRouter: User not authenticated")
		// TODO: 发送错误响应给客户端
		_ = request.GetConnection().SendMsg(protocol.MsgIDCreateGroupResp, []byte(`{"error":"用户未登录"}`))
		return
	}

	uid := userID.(uint)

	var req model.CreateGroupReq
	if err := json.Unmarshal(request.GetData(), &req); err != nil {
		fmt.Println("CreateGroupRouter: Invalid request data format - ", err)
		_ = request.GetConnection().SendMsg(protocol.MsgIDCreateGroupResp, []byte(`{"error":"请求数据格式错误"}`))
		return
	}

	createdGroup, err := global.GroupService.CreateGroup(uid, &req)
	if err != nil {
		fmt.Println("CreateGroupRouter: Failed to create group - ", err)
		respData, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("创建群组失败: %s", err.Error())})
		_ = request.GetConnection().SendMsg(protocol.MsgIDCreateGroupResp, respData)
		return
	}

	resp := model.CreateGroupResp{
		ID:          createdGroup.ID,
		Name:        createdGroup.Name,
		OwnerUserID: createdGroup.OwnerUserID,
		Description: createdGroup.Description,
		Avatar:      createdGroup.Avatar,
		MemberCount: createdGroup.MemberCount,
		CreatedAt:   createdGroup.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	respData, _ := json.Marshal(resp)
	_ = request.GetConnection().SendMsg(protocol.MsgIDCreateGroupResp, respData)
	fmt.Printf("User %d created group %s (ID: %d) successfully\n", uid, createdGroup.Name, createdGroup.ID)
}

// --- JoinGroupRouter --- //
type JoinGroupRouter struct {
	znet.BaseRouter
}

func (r *JoinGroupRouter) Handle(request ziface.IRequest) {
	userID, err := request.GetConnection().GetProperty("userID")
	if err != nil || userID == nil {
		fmt.Println("JoinGroupRouter: User not authenticated")
		_ = request.GetConnection().SendMsg(protocol.MsgIDJoinGroupResp, []byte(`{"error":"用户未登录"}`))
		return
	}
	uid := userID.(uint)

	var req model.JoinGroupReq
	if err := json.Unmarshal(request.GetData(), &req); err != nil {
		fmt.Println("JoinGroupRouter: Invalid request data format - ", err)
		_ = request.GetConnection().SendMsg(protocol.MsgIDJoinGroupResp, []byte(`{"error":"请求数据格式错误"}`))
		return
	}

	if err := global.GroupService.JoinGroup(uid, &req); err != nil {
		fmt.Printf("JoinGroupRouter: User %d failed to join group %d - %s\n", uid, req.GroupID, err.Error())
		respData, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("加入群组失败: %s", err.Error())})
		_ = request.GetConnection().SendMsg(protocol.MsgIDJoinGroupResp, respData)
		return
	}

	respData, _ := json.Marshal(map[string]string{"message": "成功加入群组"})
	_ = request.GetConnection().SendMsg(protocol.MsgIDJoinGroupResp, respData)
	fmt.Printf("User %d joined group %d successfully\n", uid, req.GroupID)
}

// --- LeaveGroupRouter --- //
type LeaveGroupRouter struct {
	znet.BaseRouter
}

func (r *LeaveGroupRouter) Handle(request ziface.IRequest) {
	userID, err := request.GetConnection().GetProperty("userID")
	if err != nil || userID == nil {
		fmt.Println("LeaveGroupRouter: User not authenticated")
		_ = request.GetConnection().SendMsg(protocol.MsgIDLeaveGroupResp, []byte(`{"error":"用户未登录"}`))
		return
	}
	uid := userID.(uint)

	var req model.LeaveGroupReq
	if err := json.Unmarshal(request.GetData(), &req); err != nil {
		fmt.Println("LeaveGroupRouter: Invalid request data format - ", err)
		_ = request.GetConnection().SendMsg(protocol.MsgIDLeaveGroupResp, []byte(`{"error":"请求数据格式错误"}`))
		return
	}

	if err := global.GroupService.LeaveGroup(uid, &req); err != nil {
		fmt.Printf("LeaveGroupRouter: User %d failed to leave group %d - %s\n", uid, req.GroupID, err.Error())
		respData, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("退出群组失败: %s", err.Error())})
		_ = request.GetConnection().SendMsg(protocol.MsgIDLeaveGroupResp, respData)
		return
	}

	respData, _ := json.Marshal(map[string]string{"message": "成功退出群组"})
	_ = request.GetConnection().SendMsg(protocol.MsgIDLeaveGroupResp, respData)
	fmt.Printf("User %d left group %d successfully\n", uid, req.GroupID)
}
