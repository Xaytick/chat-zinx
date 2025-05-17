package service

import "github.com/Xaytick/chat-zinx/chat-server/pkg/model"

// IGroupService 定义群组服务接口
type IGroupService interface {
	CreateGroup(userID uint, req *model.CreateGroupReq) (*model.Group, error)
	JoinGroup(userID uint, req *model.JoinGroupReq) error
	LeaveGroup(userID uint, req *model.LeaveGroupReq) error
	// GetGroupDetails(groupID uint) (*model.Group, error) // 暂未实现，后续可添加获取群详细信息（包括成员列表）
	// GetUserGroups(userID uint) ([]*model.GroupBasicInfo, error) // 暂未实现，后续可添加获取用户加入的群列表
}
