package mysql

import (
	"errors"
	"fmt"
	"time"

	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"gorm.io/gorm"
)

// CreateGroup 在数据库中创建群组记录，并自动添加群主为成员 (GORM实现)
func CreateGroup(group *model.Group, ownerAsMember *model.GroupMember) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		// 1. 设置群组创建时间并创建群组
		group.CreatedAt = time.Now()
		group.UpdatedAt = time.Now()
		group.MemberCount = 1 // 群主是第一个成员
		if err := tx.Create(group).Error; err != nil {
			return fmt.Errorf("failed to create group in transaction: %w", err)
		}

		// 2. 设置群主成员信息并创建
		ownerAsMember.GroupID = group.ID         // 关联到刚创建的群组ID
		ownerAsMember.UserID = group.OwnerUserID // 群主的 User ID
		ownerAsMember.Role = model.GroupRoleOwner
		ownerAsMember.JoinedAt = time.Now()
		ownerAsMember.CreatedAt = time.Now()
		ownerAsMember.UpdatedAt = time.Now()
		if err := tx.Create(ownerAsMember).Error; err != nil {
			return fmt.Errorf("failed to add owner as group member in transaction: %w", err)
		}

		return nil
	})
}

// GetGroupByID 根据群组ID从数据库中获取群组信息 (GORM实现)
func GetGroupByID(groupID uint) (*model.Group, error) {
	var group model.Group
	result := DB.First(&group, groupID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // 或者返回一个自定义的 RecordNotFoundError
		}
		return nil, result.Error
	}
	return &group, nil
}

// AddGroupMember向群组中添加成员 (GORM实现)
func AddGroupMember(member *model.GroupMember) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		member.JoinedAt = time.Now()
		member.CreatedAt = time.Now()
		member.UpdatedAt = time.Now()
		if err := tx.Create(member).Error; err != nil {
			// 可能是 uniqueIndex(idx_group_user) 冲突，即用户已在群组中
			return fmt.Errorf("failed to add group member in transaction: %w", err)
		}

		// 更新群组成员数量
		if err := tx.Model(&model.Group{}).Where("id = ?", member.GroupID).UpdateColumn("member_count", gorm.Expr("member_count + ?", 1)).Error; err != nil {
			return fmt.Errorf("failed to update group member_count in transaction: %w", err)
		}
		return nil
	})
}

// GetGroupMember 获取指定的群组成员信息 (GORM实现)
func GetGroupMember(groupID, userID uint) (*model.GroupMember, error) {
	var member model.GroupMember
	result := DB.Where("group_id = ? AND user_id = ?", groupID, userID).First(&member)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // 未找到
		}
		return nil, result.Error
	}
	return &member, nil
}

// RemoveGroupMember 从群组中移除成员 (GORM实现)
func RemoveGroupMember(groupID, userID uint) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		// 先找到要删除的成员，确保存在
		var member model.GroupMember
		findResult := tx.Where("group_id = ? AND user_id = ?", groupID, userID).First(&member)
		if findResult.Error != nil {
			if errors.Is(findResult.Error, gorm.ErrRecordNotFound) {
				return errors.New("member not found in group")
			}
			return fmt.Errorf("failed to find member to remove: %w", findResult.Error)
		}

		// 执行删除
		if err := tx.Delete(&member).Error; err != nil {
			return fmt.Errorf("failed to delete group member in transaction: %w", err)
		}

		// 更新群组成员数量
		if err := tx.Model(&model.Group{}).Where("id = ?", groupID).UpdateColumn("member_count", gorm.Expr("GREATEST(0, member_count - 1)")).Error; err != nil {
			return fmt.Errorf("failed to update group member_count after removal: %w", err)
		}
		return nil
	})
}

// GetGroupMembers 获取群组的所有成员信息 (GORM实现)
// 可以根据需要选择性加载User的关联信息: .Preload("User")
func GetGroupMembers(groupID uint) ([]*model.GroupMember, error) {
	var members []*model.GroupMember
	// 如果 GroupMember 模型中定义了与 User 的关联，可以使用 Preload
	// err := DB.Preload("User").Where("group_id = ?", groupID).Find(&members).Error
	err := DB.Where("group_id = ?", groupID).Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

// IsUserInGroup 检查用户是否在群组中 (GORM实现)
func IsUserInGroup(userID, groupID uint) (bool, error) {
	var count int64
	err := DB.Model(&model.GroupMember{}).Where("user_id = ? AND group_id = ?", userID, groupID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetUserGroups 获取用户加入的所有群组的基本信息 (GORM实现)
func GetUserGroups(userID uint) ([]*model.Group, error) {
	var groups []*model.Group
	// 使用 Join 查询 group_members 表来找到用户所在的所有群组
	// GORM 的 Joins 会自动处理好表名和字段名（如果模型定义正确）
	// SELECT groups.* FROM groups JOIN group_members ON groups.id = group_members.group_id WHERE group_members.user_id = ?
	result := DB.Joins("JOIN group_members ON groups.id = group_members.group_id").
		Where("group_members.user_id = ?", userID).
		Find(&groups)
	if result.Error != nil {
		return nil, result.Error
	}
	return groups, nil
}

// GetGroupMemberIDs 获取群组所有成员的 UserID 列表 (GORM实现)
func GetGroupMemberIDs(groupID uint) ([]uint, error) {
	var userIDs []uint
	// SELECT user_id FROM group_members WHERE group_id = ?
	err := DB.Model(&model.GroupMember{}).Where("group_id = ?", groupID).Pluck("user_id", &userIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get group member IDs: %w", err)
	}
	return userIDs, nil
}

// --- 新增DAO方法 --- //

// UpdateGroupMemberRole 更新群组成员角色 (GORM实现)
func UpdateGroupMemberRole(groupID, userID uint, newRole string) error {
	result := DB.Model(&model.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Updates(map[string]interface{}{
			"role":       newRole,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return fmt.Errorf("failed to update group member role: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("member not found or no changes made")
	}
	return nil
}

// UpdateGroupInfo 更新群组信息 (GORM实现)
func UpdateGroupInfo(groupID uint, updates map[string]interface{}) error {
	result := DB.Model(&model.Group{}).Where("id = ?", groupID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update group info: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("group not found or no changes made")
	}
	return nil
}
