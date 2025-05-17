package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/Xaytick/chat-zinx/chat-server/conf"
	"github.com/Xaytick/chat-zinx/chat-server/dao/mysql"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

// userService 结构体不需要导出
type userService struct{}

// NewMySQLUserService 创建一个新的用户服务实例 (MySQL实现)
// 函数名更改为 NewMySQLUserService 以明确其实现方式
func NewMySQLUserService() IUserService { // 保持 NewUserService，外部通过接口调用
	return &userService{}
}

// Register 处理用户注册逻辑
func (s *userService) Register(req *model.UserRegisterReq) (*model.User, error) {
	// 1. 检查用户名是否已存在
	existingUser, err := mysql.GetUserByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username existence: %w", err)
	}
	if existingUser != nil {
		return nil, ErrUsernameExists // Use defined error
	}

	// 2. 准备 User 模型
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		// Password 会在 DAO 层被哈希
		// UserUUID 会在 DAO 层生成
	}

	// 3. 创建用户 (DAO层负责哈希密码和生成UserUUID)
	if err := mysql.CreateUser(user, req.Password); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil // 返回创建后的用户信息（包含ID, UserUUID等）
}

// Login 处理用户登录逻辑
func (s *userService) Login(req *model.UserLoginReq) (string, *model.User, error) {
	// 1. 根据用户名获取用户信息
	user, err := mysql.GetUserByUsername(req.Username)
	if err != nil {
		// Log the detailed error, but return a generic one to the client
		fmt.Printf("Error in Login - GetUserByUsername for '%s': %v\n", req.Username, err)
		return "", nil, ErrInvalidCredentials // Generic error for security
	}
	if user == nil {
		return "", nil, ErrUserNotFound
	}

	// 2. 验证密码 (user.Password 是哈希后的密码)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", nil, ErrPasswordIncorrect
		}
		fmt.Printf("Error in Login - Password comparison for user '%s': %v\n", req.Username, err)
		return "", nil, ErrInvalidCredentials // Generic error for other bcrypt issues
	}

	// 3. 更新最后登录信息和在线状态
	if err := s.UpdateUserLastLoginInfo(user.ID); err != nil {
		fmt.Printf("Warning: failed to update user last login info for user %d: %v\n", user.ID, err)
		// Non-critical error, proceed with login
	}
	if err := s.UpdateUserOnlineStatus(user.ID, true); err != nil {
		fmt.Printf("Warning: failed to update user online status for user %d: %v\n", user.ID, err)
		// Non-critical error, proceed with login
	}

	// 4. 生成JWT Token
	jwtConf := conf.GetAuthConfig().JWT
	claims := model.CustomClaims{
		ID:       user.ID,
		UserUUID: user.UserUUID,
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * time.Duration(jwtConf.ExpiresIn)).Unix(),
			Issuer:    jwtConf.Issuer,
			NotBefore: time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtConf.Secret))
	if err != nil {
		return "", nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, user, nil
}

// GetUserByID 根据用户ID获取用户信息
func (s *userService) GetUserByID(userID uint) (*model.User, error) {
	user, err := mysql.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, mysql.ErrRecordNotFound) { // Assuming dao exports ErrRecordNotFound
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id from service: %w", err)
	}
	// mysql.GetUserByID should return nil, nil if not found and no other error occurred,
	// or an error (like ErrRecordNotFound)
	if user == nil { // This case might be redundant if ErrRecordNotFound is handled
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetUserByUUID 根据用户UUID获取用户信息
func (s *userService) GetUserByUUID(uuid string) (*model.User, error) {
	user, err := mysql.GetUserByUUID(uuid) // Assumes mysql.GetUserByUUID will be implemented
	if err != nil {
		if errors.Is(err, mysql.ErrRecordNotFound) { // Assuming dao exports ErrRecordNotFound
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by uuid from service: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetUserByUsername 根据用户名获取用户信息
func (s *userService) GetUserByUsername(username string) (*model.User, error) {
	user, err := mysql.GetUserByUsername(username) // Assumes mysql.GetUserByUsername is implemented
	if err != nil {
		if errors.Is(err, mysql.ErrRecordNotFound) { // Assuming dao exports ErrRecordNotFound
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username from service: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateUserOnlineStatus 更新用户在线状态
func (s *userService) UpdateUserOnlineStatus(userID uint, isOnline bool) error {
	return mysql.UpdateUserOnlineStatus(userID, isOnline) // Assumes mysql.UpdateUserOnlineStatus will be implemented
}

// UpdateUserLastLoginInfo 更新用户最后登录IP和时间
func (s *userService) UpdateUserLastLoginInfo(userID uint) error {
	// Potentially add clientIP if available/needed
	return mysql.UpdateUserLastLoginInfo(userID) // Assumes mysql.UpdateUserLastLoginInfo will be implemented
}
