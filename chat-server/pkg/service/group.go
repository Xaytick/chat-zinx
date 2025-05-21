package service

import "github.com/Xaytick/chat-zinx/chat-server/pkg/model"

// IGroupService 定义群组服务接口
type IGroupService interface {
	CreateGroup(userID uint, req *model.CreateGroupReq) (*model.Group, error)
	JoinGroup(userID uint, req *model.JoinGroupReq) error
	LeaveGroup(userID uint, req *model.LeaveGroupReq) error

	// Methods for group messaging and info retrieval
	IsUserInGroup(userID uint, groupID uint) (bool, error)
	GetGroupMemberIDs(groupID uint) ([]uint, error)

	// 获取群组详情 - 新增
	GetGroupDetails(groupID uint) (*model.Group, error)
	// 获取用户加入的群列表 - 新增
	GetUserGroups(userID uint) ([]*model.Group, error)
	// 获取群组成员列表 - 新增
	GetGroupMembers(groupID uint) ([]*model.GroupMember, error)
	// 获取群组成员详细信息(包含用户信息) - 新增
	GetGroupMembersWithUserInfo(groupID uint) ([]*model.GroupMemberInfo, error)

	// 群组管理相关 - 新增
	SetGroupMemberRole(operatorID uint, groupID uint, targetUserID uint, newRole string) error
	RemoveMemberFromGroup(operatorID uint, groupID uint, targetUserID uint) error
	UpdateGroupInfo(operatorID uint, groupID uint, updateReq *model.UpdateGroupInfoReq) error
}
