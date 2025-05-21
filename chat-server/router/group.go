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

// --- GetUserGroupsRouter 获取用户加入的群组列表 --- //
type GetUserGroupsRouter struct {
	znet.BaseRouter
}

func (r *GetUserGroupsRouter) Handle(request ziface.IRequest) {
	userID, err := request.GetConnection().GetProperty("userID")
	if err != nil || userID == nil {
		fmt.Println("GetUserGroupsRouter: User not authenticated")
		_ = request.GetConnection().SendMsg(protocol.MsgIDGetUserGroupsResp, []byte(`{"error":"用户未登录"}`))
		return
	}
	uid := userID.(uint)

	// 获取用户的群组列表
	groups, err := global.GroupService.GetUserGroups(uid)
	if err != nil {
		fmt.Printf("GetUserGroupsRouter: Failed to get groups for user %d - %s\n", uid, err.Error())
		respData, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("获取群组列表失败: %s", err.Error())})
		_ = request.GetConnection().SendMsg(protocol.MsgIDGetUserGroupsResp, respData)
		return
	}

	// 转换为响应格式
	groupBasicInfos := make([]*model.GroupBasicInfo, 0, len(groups))
	for _, group := range groups {
		groupBasicInfos = append(groupBasicInfos, &model.GroupBasicInfo{
			ID:          group.ID,
			Name:        group.Name,
			MemberCount: group.MemberCount,
			Description: group.Description,
		})
	}

	resp := model.GetUserGroupsResp{
		Groups: groupBasicInfos,
		Total:  len(groupBasicInfos),
	}

	respData, _ := json.Marshal(resp)
	_ = request.GetConnection().SendMsg(protocol.MsgIDGetUserGroupsResp, respData)
	fmt.Printf("Retrieved %d groups for user %d\n", len(groups), uid)
}

// --- GetGroupMembersRouter 获取群组成员列表 --- //
type GetGroupMembersRouter struct {
	znet.BaseRouter
}

func (r *GetGroupMembersRouter) Handle(request ziface.IRequest) {
	userID, err := request.GetConnection().GetProperty("userID")
	if err != nil || userID == nil {
		fmt.Println("GetGroupMembersRouter: User not authenticated")
		_ = request.GetConnection().SendMsg(protocol.MsgIDGetGroupMembersResp, []byte(`{"error":"用户未登录"}`))
		return
	}
	uid := userID.(uint)

	var req model.GetGroupMembersReq
	if err := json.Unmarshal(request.GetData(), &req); err != nil {
		fmt.Println("GetGroupMembersRouter: Invalid request data format - ", err)
		_ = request.GetConnection().SendMsg(protocol.MsgIDGetGroupMembersResp, []byte(`{"error":"请求数据格式错误"}`))
		return
	}

	// 检查用户是否是群成员
	isMember, err := global.GroupService.IsUserInGroup(uid, req.GroupID)
	if err != nil {
		fmt.Printf("GetGroupMembersRouter: Failed to check membership for user %d in group %d - %s\n", uid, req.GroupID, err.Error())
		respData, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("检查群组成员资格失败: %s", err.Error())})
		_ = request.GetConnection().SendMsg(protocol.MsgIDGetGroupMembersResp, respData)
		return
	}

	if !isMember {
		fmt.Printf("GetGroupMembersRouter: User %d is not a member of group %d\n", uid, req.GroupID)
		respData, _ := json.Marshal(map[string]string{"error": "您不是该群组的成员"})
		_ = request.GetConnection().SendMsg(protocol.MsgIDGetGroupMembersResp, respData)
		return
	}

	// 获取群组成员详细信息
	members, err := global.GroupService.GetGroupMembersWithUserInfo(req.GroupID)
	if err != nil {
		fmt.Printf("GetGroupMembersRouter: Failed to get members for group %d - %s\n", req.GroupID, err.Error())
		respData, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("获取群组成员失败: %s", err.Error())})
		_ = request.GetConnection().SendMsg(protocol.MsgIDGetGroupMembersResp, respData)
		return
	}

	resp := model.GetGroupMembersResp{
		GroupID: req.GroupID,
		Members: members,
		Total:   len(members),
	}

	respData, _ := json.Marshal(resp)
	_ = request.GetConnection().SendMsg(protocol.MsgIDGetGroupMembersResp, respData)
	fmt.Printf("Retrieved %d members for group %d\n", len(members), req.GroupID)
}

// --- GetGroupDetailsRouter 获取群组详情 --- //
type GetGroupDetailsRouter struct {
	znet.BaseRouter
}

func (r *GetGroupDetailsRouter) Handle(request ziface.IRequest) {
	userID, err := request.GetConnection().GetProperty("userID")
	if err != nil || userID == nil {
		fmt.Println("GetGroupDetailsRouter: User not authenticated")
		_ = request.GetConnection().SendMsg(protocol.MsgIDGetGroupDetailsResp, []byte(`{"error":"用户未登录"}`))
		return
	}
	uid := userID.(uint)

	// 解析请求
	type GetGroupDetailsReq struct {
		GroupID uint `json:"group_id"`
	}
	var req GetGroupDetailsReq
	if err := json.Unmarshal(request.GetData(), &req); err != nil {
		fmt.Println("GetGroupDetailsRouter: Invalid request data format - ", err)
		_ = request.GetConnection().SendMsg(protocol.MsgIDGetGroupDetailsResp, []byte(`{"error":"请求数据格式错误"}`))
		return
	}

	// 检查用户是否是群成员
	isMember, err := global.GroupService.IsUserInGroup(uid, req.GroupID)
	if err != nil {
		fmt.Printf("GetGroupDetailsRouter: Failed to check membership for user %d in group %d - %s\n", uid, req.GroupID, err.Error())
		respData, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("检查群组成员资格失败: %s", err.Error())})
		_ = request.GetConnection().SendMsg(protocol.MsgIDGetGroupDetailsResp, respData)
		return
	}

	if !isMember {
		fmt.Printf("GetGroupDetailsRouter: User %d is not a member of group %d\n", uid, req.GroupID)
		respData, _ := json.Marshal(map[string]string{"error": "您不是该群组的成员"})
		_ = request.GetConnection().SendMsg(protocol.MsgIDGetGroupDetailsResp, respData)
		return
	}

	// 获取群组详情
	group, err := global.GroupService.GetGroupDetails(req.GroupID)
	if err != nil {
		fmt.Printf("GetGroupDetailsRouter: Failed to get details for group %d - %s\n", req.GroupID, err.Error())
		respData, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("获取群组详情失败: %s", err.Error())})
		_ = request.GetConnection().SendMsg(protocol.MsgIDGetGroupDetailsResp, respData)
		return
	}

	if group == nil {
		fmt.Printf("GetGroupDetailsRouter: Group %d not found\n", req.GroupID)
		respData, _ := json.Marshal(map[string]string{"error": "群组不存在"})
		_ = request.GetConnection().SendMsg(protocol.MsgIDGetGroupDetailsResp, respData)
		return
	}

	// 构造响应
	type GroupDetailsResp struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		OwnerUserID uint   `json:"owner_user_id"`
		Description string `json:"description"`
		Avatar      string `json:"avatar"`
		MemberCount uint   `json:"member_count"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
	}

	resp := GroupDetailsResp{
		ID:          group.ID,
		Name:        group.Name,
		OwnerUserID: group.OwnerUserID,
		Description: group.Description,
		Avatar:      group.Avatar,
		MemberCount: group.MemberCount,
		CreatedAt:   group.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   group.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	respData, _ := json.Marshal(resp)
	_ = request.GetConnection().SendMsg(protocol.MsgIDGetGroupDetailsResp, respData)
	fmt.Printf("Retrieved details for group %d\n", req.GroupID)
}

// --- UpdateGroupInfoRouter 更新群组信息 --- //
type UpdateGroupInfoRouter struct {
	znet.BaseRouter
}

func (r *UpdateGroupInfoRouter) Handle(request ziface.IRequest) {
	userID, err := request.GetConnection().GetProperty("userID")
	if err != nil || userID == nil {
		fmt.Println("UpdateGroupInfoRouter: User not authenticated")
		_ = request.GetConnection().SendMsg(protocol.MsgIDUpdateGroupInfoResp, []byte(`{"error":"用户未登录"}`))
		return
	}
	uid := userID.(uint)

	var req model.UpdateGroupInfoReq
	if err := json.Unmarshal(request.GetData(), &req); err != nil {
		fmt.Println("UpdateGroupInfoRouter: Invalid request data format - ", err)
		_ = request.GetConnection().SendMsg(protocol.MsgIDUpdateGroupInfoResp, []byte(`{"error":"请求数据格式错误"}`))
		return
	}

	// 更新群组信息
	if err := global.GroupService.UpdateGroupInfo(uid, req.GroupID, &req); err != nil {
		fmt.Printf("UpdateGroupInfoRouter: User %d failed to update group %d - %s\n", uid, req.GroupID, err.Error())
		respData, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("更新群组信息失败: %s", err.Error())})
		_ = request.GetConnection().SendMsg(protocol.MsgIDUpdateGroupInfoResp, respData)
		return
	}

	respData, _ := json.Marshal(map[string]string{"message": "群组信息更新成功"})
	_ = request.GetConnection().SendMsg(protocol.MsgIDUpdateGroupInfoResp, respData)
	fmt.Printf("User %d updated info for group %d successfully\n", uid, req.GroupID)
}

// --- SetMemberRoleRouter 设置群成员角色 --- //
type SetMemberRoleRouter struct {
	znet.BaseRouter
}

func (r *SetMemberRoleRouter) Handle(request ziface.IRequest) {
	userID, err := request.GetConnection().GetProperty("userID")
	if err != nil || userID == nil {
		fmt.Println("SetMemberRoleRouter: User not authenticated")
		_ = request.GetConnection().SendMsg(protocol.MsgIDSetMemberRoleResp, []byte(`{"error":"用户未登录"}`))
		return
	}
	uid := userID.(uint)

	var req model.SetGroupMemberRoleReq
	if err := json.Unmarshal(request.GetData(), &req); err != nil {
		fmt.Println("SetMemberRoleRouter: Invalid request data format - ", err)
		_ = request.GetConnection().SendMsg(protocol.MsgIDSetMemberRoleResp, []byte(`{"error":"请求数据格式错误"}`))
		return
	}

	// 设置成员角色
	if err := global.GroupService.SetGroupMemberRole(uid, req.GroupID, req.TargetUserID, req.NewRole); err != nil {
		fmt.Printf("SetMemberRoleRouter: User %d failed to set role for user %d in group %d - %s\n",
			uid, req.TargetUserID, req.GroupID, err.Error())
		respData, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("设置成员角色失败: %s", err.Error())})
		_ = request.GetConnection().SendMsg(protocol.MsgIDSetMemberRoleResp, respData)
		return
	}

	respData, _ := json.Marshal(map[string]string{"message": "成员角色设置成功"})
	_ = request.GetConnection().SendMsg(protocol.MsgIDSetMemberRoleResp, respData)
	fmt.Printf("User %d set role for user %d in group %d to %s successfully\n",
		uid, req.TargetUserID, req.GroupID, req.NewRole)
}

// --- RemoveMemberRouter 将成员移出群组 --- //
type RemoveMemberRouter struct {
	znet.BaseRouter
}

func (r *RemoveMemberRouter) Handle(request ziface.IRequest) {
	userID, err := request.GetConnection().GetProperty("userID")
	if err != nil || userID == nil {
		fmt.Println("RemoveMemberRouter: User not authenticated")
		_ = request.GetConnection().SendMsg(protocol.MsgIDRemoveMemberResp, []byte(`{"error":"用户未登录"}`))
		return
	}
	uid := userID.(uint)

	var req model.RemoveMemberReq
	if err := json.Unmarshal(request.GetData(), &req); err != nil {
		fmt.Println("RemoveMemberRouter: Invalid request data format - ", err)
		_ = request.GetConnection().SendMsg(protocol.MsgIDRemoveMemberResp, []byte(`{"error":"请求数据格式错误"}`))
		return
	}

	// 移除成员
	if err := global.GroupService.RemoveMemberFromGroup(uid, req.GroupID, req.TargetUserID); err != nil {
		fmt.Printf("RemoveMemberRouter: User %d failed to remove user %d from group %d - %s\n",
			uid, req.TargetUserID, req.GroupID, err.Error())
		respData, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("移除成员失败: %s", err.Error())})
		_ = request.GetConnection().SendMsg(protocol.MsgIDRemoveMemberResp, respData)
		return
	}

	respData, _ := json.Marshal(map[string]string{"message": "成员已从群组中移除"})
	_ = request.GetConnection().SendMsg(protocol.MsgIDRemoveMemberResp, respData)
	fmt.Printf("User %d removed user %d from group %d successfully\n", uid, req.TargetUserID, req.GroupID)
}
