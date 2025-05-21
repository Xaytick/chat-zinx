package model

import "time"

// Group 群组信息
type Group struct {
	ID          uint      `json:"id" gorm:"primarykey"` // GORM 默认使用 ID 作为主键
	Name        string    `json:"name" gorm:"type:varchar(100);not null;uniqueIndex:idx_group_name"`
	OwnerUserID uint      `json:"owner_user_id" gorm:"not null;index"` // 群主的用户ID (关联User表的ID)
	Avatar      string    `json:"avatar" gorm:"type:varchar(255)"`
	Description string    `json:"description" gorm:"type:varchar(500)"`
	MemberCount uint      `json:"member_count" gorm:"default:1"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GroupMember 群组成员信息
type GroupMember struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	GroupID   uint      `json:"group_id" gorm:"not null;uniqueIndex:idx_group_user"` // 外键，关联 Group 表的 ID
	UserID    uint      `json:"user_id" gorm:"not null;uniqueIndex:idx_group_user"`  // 外键，关联 User 表的 ID
	Role      string    `json:"role" gorm:"type:varchar(20);default:'member'"`       // e.g., "owner", "admin", "member"
	JoinedAt  time.Time `json:"joined_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Constants for GroupMember Role
const (
	GroupRoleOwner  = "owner"
	GroupRoleAdmin  = "admin"
	GroupRoleMember = "member"
)

// --- Request and Response Structs ---

// CreateGroupReq 创建群组请求
type CreateGroupReq struct {
	Name        string `json:"name" binding:"required,min=2,max=30"`
	Description string `json:"description" binding:"max=200"`
	Avatar      string `json:"avatar"`
}

// CreateGroupResp 创建群组响应
type CreateGroupResp struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	OwnerUserID uint   `json:"owner_user_id"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	MemberCount uint   `json:"member_count"`
	CreatedAt   string `json:"created_at"`
}

// JoinGroupReq 加入群组请求
type JoinGroupReq struct {
	GroupID uint `json:"group_id" binding:"required"`
}

// LeaveGroupReq 退出群组请求
type LeaveGroupReq struct {
	GroupID uint `json:"group_id" binding:"required"`
}

// GroupBasicInfo 群组基本信息，用于列表等场景
type GroupBasicInfo struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	MemberCount uint   `json:"member_count"`
	Description string `json:"description"`
}

// 新增模型定义

// GroupMemberInfo 群组成员信息（包含用户信息）
type GroupMemberInfo struct {
	MemberID uint      `json:"member_id"` // GroupMember表中的ID
	UserID   uint      `json:"user_id"`   // 用户ID
	UserUUID string    `json:"user_uuid"` // 用户UUID
	Username string    `json:"username"`  // 用户名
	Role     string    `json:"role"`      // 在群组中的角色
	JoinedAt time.Time `json:"joined_at"` // 加入时间
	IsOnline bool      `json:"is_online"` // 在线状态
}

// GetGroupMembersReq 获取群成员列表请求
type GetGroupMembersReq struct {
	GroupID uint `json:"group_id" binding:"required"`
}

// GetGroupMembersResp 获取群成员列表响应
type GetGroupMembersResp struct {
	GroupID uint               `json:"group_id"`
	Members []*GroupMemberInfo `json:"members"`
	Total   int                `json:"total"`
}

// GetUserGroupsReq 获取用户加入的群列表请求
type GetUserGroupsReq struct {
	// 可选字段，如过滤条件等
}

// GetUserGroupsResp 获取用户加入的群列表响应
type GetUserGroupsResp struct {
	Groups []*GroupBasicInfo `json:"groups"`
	Total  int               `json:"total"`
}

// UpdateGroupInfoReq 更新群组信息请求
type UpdateGroupInfoReq struct {
	GroupID     uint   `json:"group_id" binding:"required"`
	Name        string `json:"name,omitempty" binding:"omitempty,min=2,max=30"`
	Description string `json:"description,omitempty" binding:"omitempty,max=200"`
	Avatar      string `json:"avatar,omitempty"`
}

// SetGroupMemberRoleReq 设置群组成员角色请求
type SetGroupMemberRoleReq struct {
	GroupID      uint   `json:"group_id" binding:"required"`
	TargetUserID uint   `json:"target_user_id" binding:"required"`
	NewRole      string `json:"new_role" binding:"required,oneof=admin member"`
}

// RemoveMemberReq 将成员移出群组请求
type RemoveMemberReq struct {
	GroupID      uint `json:"group_id" binding:"required"`
	TargetUserID uint `json:"target_user_id" binding:"required"`
}
