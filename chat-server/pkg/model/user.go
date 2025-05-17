package model

import (
	"time"
)

// User 用户模型，包含用户基本信息
type User struct {
	ID        uint      `json:"id" gorm:"primarykey"`                                   // GORM 主键
	UserUUID  string    `json:"user_uuid" gorm:"type:varchar(36);uniqueIndex;not null"` // 用户唯一业务标识 (替代原 UserID)
	Username  string    `json:"username" gorm:"type:varchar(50);uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"type:varchar(100);not null"` // 存储哈希后的密码, JSON中忽略
	Email     string    `json:"email" gorm:"type:varchar(100);uniqueIndex"`
	Avatar    string    `json:"avatar" gorm:"type:varchar(255)"` // 新增头像字段
	Gender    string    `json:"gender" gorm:"type:varchar(10)"`  // 新增性别字段
	IsOnline  bool      `json:"is_online" gorm:"default:false"`  // Added IsOnline field
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	LastLogin time.Time `json:"last_login,omitempty"`
}

// UserRegisterReq 用户注册请求结构
type UserRegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	Email    string `json:"email" binding:"required,email"`
}

// UserLoginReq 用户登录请求结构
type UserLoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserBasicInfo 用户基本信息，用于列表或嵌入其他响应中
type UserBasicInfo struct {
	ID       uint   `json:"id"`
	UserUUID string `json:"user_uuid"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

// UserLoginResponse 用户登录响应结构
type UserLoginResponse struct {
	ID        uint      `json:"id"`
	UserUUID  string    `json:"user_uuid"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar"`
	LastLogin time.Time `json:"last_login"`
	Token     string    `json:"token"`
}

// UserRegisterResponse 用户注册响应结构 (通常注册成功后直接返回用户信息和Token，类似登录响应)
type UserRegisterResponse struct {
	ID       uint   `json:"id"`
	UserUUID string `json:"user_uuid"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Token    string `json:"token,omitempty"` // 注册后也可以选择不立即返回token，让用户去登录
}
