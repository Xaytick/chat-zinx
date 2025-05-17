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
	Avatar      string `json:"avatar"`
	MemberCount uint   `json:"member_count"`
	Description string `json:"description"`
}
