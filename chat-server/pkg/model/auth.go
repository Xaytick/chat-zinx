package model

import "github.com/golang-jwt/jwt"

// CustomClaims 自定义JWT Claims，包含用户基本信息
type CustomClaims struct {
	ID       uint   `json:"id"`        // 用户主键ID
	UserUUID string `json:"user_uuid"` // 用户业务UUID
	Username string `json:"username"`
	jwt.StandardClaims
}
