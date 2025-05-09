package model

import (
	"time"
)

// User 用户模型，包含用户基本信息
type User struct {
	UserID       string    `json:"user_id"`       // 用户唯一标识
	Username     string    `json:"username"`      // 用户名
	PasswordHash string    `json:"password_hash"` // 密码哈希(不存明文)
	Salt         string    `json:"salt"`          // 密码盐(用于加密)
	Email        string    `json:"email"`         // 邮箱
	CreatedAt    time.Time `json:"created_at"`    // 注册时间
	UpdatedAt    time.Time `json:"updated_at"`    // 更新时间
	LastLogin    time.Time `json:"last_login"`    // 最后登录时间
}

// UserRegisterReq 用户注册请求结构
type UserRegisterReq struct {
	Username string `json:"username"` // 用户名
	Password string `json:"password"` // 密码
	Email    string `json:"email"`    // 邮箱
}

// UserLoginReq 用户登录请求结构
type UserLoginReq struct {
	Username string `json:"username"` // 用户名
	Password string `json:"password"` // 密码
}

// UserResponse 用户信息响应结构(不含敏感信息)
type UserRegisterResponse struct {
	UserID    string `json:"user_id"`    // 用户ID
	Username  string `json:"username"`   // 用户名
	Email     string `json:"email"`      // 邮箱
}

// UserLoginResponse 用户登录响应结构
type UserLoginResponse struct {
	UserID    string `json:"user_id"`    // 用户ID
	Username  string `json:"username"`   // 用户名
	Email     string `json:"email"`      // 邮箱
	LastLogin string `json:"last_login"` // 最后登录时间
}
