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

// IsUserInGroup 检查用户是否在指定的群组中
func (s *groupService) IsUserInGroup(userID uint, groupID uint) (bool, error) {
	return mysql.IsUserInGroup(userID, groupID)
}

// GetGroupMemberIDs 获取指定群组的所有成员 UserID 列表
func (s *groupService) GetGroupMemberIDs(groupID uint) ([]uint, error) {
	return mysql.GetGroupMemberIDs(groupID)
}

// --- 新增方法实现 --- //

// GetGroupDetails 获取群组详情
func (s *groupService) GetGroupDetails(groupID uint) (*model.Group, error) {
	return mysql.GetGroupByID(groupID)
}

// GetUserGroups 获取用户加入的所有群组
func (s *groupService) GetUserGroups(userID uint) ([]*model.Group, error) {
	return mysql.GetUserGroups(userID)
}

// GetGroupMembers 获取群组所有成员
func (s *groupService) GetGroupMembers(groupID uint) ([]*model.GroupMember, error) {
	return mysql.GetGroupMembers(groupID)
}

// GetGroupMembersWithUserInfo 获取群组成员详细信息（包含用户信息）
func (s *groupService) GetGroupMembersWithUserInfo(groupID uint) ([]*model.GroupMemberInfo, error) {
	// 1. 获取群组成员
	members, err := mysql.GetGroupMembers(groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}

	if len(members) == 0 {
		return []*model.GroupMemberInfo{}, nil
	}

	// 2. 收集所有用户ID
	userIDs := make([]uint, 0, len(members))
	for _, member := range members {
		userIDs = append(userIDs, member.UserID)
	}

	// 3. 批量获取用户信息
	users, err := mysql.GetUsersByIDs(userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get users info: %w", err)
	}

	// 4. 建立用户ID到用户信息的映射，方便快速查找
	userMap := make(map[uint]*model.User, len(users))
	for _, user := range users {
		userMap[user.ID] = user
	}

	// 5. 构建结果
	result := make([]*model.GroupMemberInfo, 0, len(members))
	for _, member := range members {
		user, exists := userMap[member.UserID]
		if !exists {
			// 如果找不到用户信息，跳过该成员
			fmt.Printf("Warning: User %d not found for group member\n", member.UserID)
			continue
		}

		memberInfo := &model.GroupMemberInfo{
			MemberID: member.ID,
			UserID:   user.ID,
			UserUUID: user.UserUUID,
			Username: user.Username,
			Role:     member.Role,
			JoinedAt: member.JoinedAt,
			IsOnline: user.IsOnline,
		}
		result = append(result, memberInfo)
	}

	return result, nil
}

// SetGroupMemberRole 设置群组成员角色
func (s *groupService) SetGroupMemberRole(operatorID uint, groupID uint, targetUserID uint, newRole string) error {
	// 1. 验证群组存在
	group, err := mysql.GetGroupByID(groupID)
	if err != nil {
		return fmt.Errorf("failed to get group: %w", err)
	}
	if group == nil {
		return errors.New("group not found")
	}

	// 2. 检查权限：只有群主可以设置管理员
	if group.OwnerUserID != operatorID {
		return errors.New("permission denied: only group owner can change member roles")
	}

	// 3. 检查目标用户是否为群成员
	targetMember, err := mysql.GetGroupMember(groupID, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to get target member: %w", err)
	}
	if targetMember == nil {
		return errors.New("target user is not a member of this group")
	}

	// 4. 检查是否试图修改群主角色
	if targetUserID == group.OwnerUserID {
		return errors.New("cannot change role of the group owner")
	}

	// 5. 验证角色有效性
	if newRole != model.GroupRoleAdmin && newRole != model.GroupRoleMember {
		return errors.New("invalid role: must be 'admin' or 'member'")
	}

	// 6. 更新角色
	return mysql.UpdateGroupMemberRole(groupID, targetUserID, newRole)
}

// RemoveMemberFromGroup 将成员移出群组
func (s *groupService) RemoveMemberFromGroup(operatorID uint, groupID uint, targetUserID uint) error {
	// 1. 验证群组存在
	group, err := mysql.GetGroupByID(groupID)
	if err != nil {
		return fmt.Errorf("failed to get group: %w", err)
	}
	if group == nil {
		return errors.New("group not found")
	}

	// 2. 检查操作者是否有权限（群主可以移除任何人，管理员可以移除普通成员）
	if operatorID != group.OwnerUserID {
		// 如果不是群主，检查是否是管理员
		operatorMember, err := mysql.GetGroupMember(groupID, operatorID)
		if err != nil {
			return fmt.Errorf("failed to get operator's membership: %w", err)
		}
		if operatorMember == nil || operatorMember.Role != model.GroupRoleAdmin {
			return errors.New("permission denied: only group owner and admins can remove members")
		}

		// 管理员不能移除其他管理员或群主
		targetMember, err := mysql.GetGroupMember(groupID, targetUserID)
		if err != nil {
			return fmt.Errorf("failed to get target member: %w", err)
		}
		if targetMember == nil {
			return errors.New("target user is not a member of this group")
		}
		if targetMember.Role == model.GroupRoleAdmin || targetUserID == group.OwnerUserID {
			return errors.New("permission denied: admins cannot remove other admins or the owner")
		}
	}

	// 3. 不能移除自己（应该使用 LeaveGroup 接口）
	if operatorID == targetUserID {
		return errors.New("cannot remove yourself from group, use leave group instead")
	}

	// 4. 不能移除群主
	if targetUserID == group.OwnerUserID {
		return errors.New("cannot remove the group owner")
	}

	// 5. 执行移除操作
	return mysql.RemoveGroupMember(groupID, targetUserID)
}

// UpdateGroupInfo 更新群组信息
func (s *groupService) UpdateGroupInfo(operatorID uint, groupID uint, updateReq *model.UpdateGroupInfoReq) error {
	// 1. 验证群组存在
	group, err := mysql.GetGroupByID(groupID)
	if err != nil {
		return fmt.Errorf("failed to get group: %w", err)
	}
	if group == nil {
		return errors.New("group not found")
	}

	// 2. 检查权限：只有群主和管理员可以更新群组信息
	if operatorID != group.OwnerUserID {
		// 如果不是群主，检查是否是管理员
		operatorMember, err := mysql.GetGroupMember(groupID, operatorID)
		if err != nil {
			return fmt.Errorf("failed to get operator's membership: %w", err)
		}
		if operatorMember == nil || operatorMember.Role != model.GroupRoleAdmin {
			return errors.New("permission denied: only group owner and admins can update group info")
		}
	}

	// 3. 更新群组信息
	updates := make(map[string]interface{})
	if updateReq.Name != "" {
		updates["name"] = updateReq.Name
	}
	if updateReq.Description != "" {
		updates["description"] = updateReq.Description
	}
	if updateReq.Avatar != "" {
		updates["avatar"] = updateReq.Avatar
	}

	if len(updates) == 0 {
		return errors.New("no updates provided")
	}

	updates["updated_at"] = time.Now()

	return mysql.UpdateGroupInfo(groupID, updates)
}
