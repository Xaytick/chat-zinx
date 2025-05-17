package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/Xaytick/chat-zinx/chat-server/dao/mysql"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
)

type groupService struct{}

// NewGroupService 创建一个新的群组服务实例
func NewGroupService() IGroupService {
	return &groupService{}
}

// CreateGroup 创建群组
func (s *groupService) CreateGroup(userID uint, req *model.CreateGroupReq) (*model.Group, error) {
	// 检查用户是否存在 (如果需要)
	// _, err := mysql.GetUserByID(userID)
	// if err != nil {
	// 	 return nil, fmt.Errorf("owner user not found: %w", err)
	// }

	group := &model.Group{
		Name:        req.Name,
		OwnerUserID: userID,
		Avatar:      req.Avatar, // 可能需要处理默认头像
		Description: req.Description,
		// MemberCount 默认为1，在DAO层设置
	}

	ownerAsMember := &model.GroupMember{}

	err := mysql.CreateGroup(group, ownerAsMember) // DAO层会处理群组创建和群主成员的添加
	if err != nil {
		// TODO: 更细致的错误处理，例如群名是否已存在 (数据库层面通过unique约束保证，这里可以转换为更友好的错误信息)
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	return group, nil
}

// JoinGroup 加入群组
func (s *groupService) JoinGroup(userID uint, req *model.JoinGroupReq) error {
	// 1. 检查群组是否存在
	group, err := mysql.GetGroupByID(req.GroupID)
	if err != nil {
		return fmt.Errorf("failed to get group: %w", err)
	}
	if group == nil {
		return errors.New("group not found")
	}

	// 2. 检查用户是否已在该群组中
	existingMember, err := mysql.GetGroupMember(req.GroupID, userID)
	if err != nil {
		return fmt.Errorf("failed to check group member: %w", err)
	}
	if existingMember != nil {
		return errors.New("user already in this group")
	}

	// 3. 添加成员
	newMember := &model.GroupMember{
		GroupID:  req.GroupID,
		UserID:   userID,
		Role:     model.GroupRoleMember, // 默认角色为普通成员
		JoinedAt: time.Now(),
	}
	if err := mysql.AddGroupMember(newMember); err != nil {
		return fmt.Errorf("failed to add user to group: %w", err)
	}
	return nil
}

// LeaveGroup 退出群组
func (s *groupService) LeaveGroup(userID uint, req *model.LeaveGroupReq) error {
	// 1. 检查群组是否存在
	group, err := mysql.GetGroupByID(req.GroupID)
	if err != nil {
		return fmt.Errorf("failed to get group info: %w", err)
	}
	if group == nil {
		return errors.New("group not found")
	}

	// 2. 检查用户是否是群主
	if group.OwnerUserID == userID {
		// TODO: 群主退群的逻辑需要更复杂处理，例如：
		// 1. 如果群内只有群主一人，可以直接解散群组。
		// 2. 如果群内还有其他成员，需要先转让群主身份，或者不允许直接退出，提示先转让。
		// 此处暂时简化为不允许群主直接退出。
		return errors.New("group owner cannot leave the group directly, please transfer ownership first")
	}

	// 3. 检查用户是否在群组中
	existingMember, err := mysql.GetGroupMember(req.GroupID, userID)
	if err != nil {
		return fmt.Errorf("failed to check group member existence: %w", err)
	}
	if existingMember == nil {
		return errors.New("user is not a member of this group")
	}

	// 4. 移除成员
	if err := mysql.RemoveGroupMember(req.GroupID, userID); err != nil {
		return fmt.Errorf("failed to remove user from group: %w", err)
	}
	return nil
}
