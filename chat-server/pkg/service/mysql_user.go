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
		// 只有当错误不是 "record not found" 时，才认为是检查过程的失败
		// mysql.ErrRecordNotFound 是 gorm.ErrRecordNotFound 的别名，在 dao/mysql/user.go 中定义并导出
		if !errors.Is(err, mysql.ErrRecordNotFound) {
			return nil, fmt.Errorf("检查用户名是否存在失败: %w", err)
		}
		// 如果错误是 "record not found"，existingUser 会是 nil (或者说 GetUserByUsername 在 record not found 时应确保返回 nil user)
		// 清除 "record not found" 错误，因为这意味着用户名可用。
		err = nil
	}

	// 此时，如果原来的 err 不是 nil 且不是 ErrRecordNotFound，上面已经返回了。
	// 如果原来的 err 是 ErrRecordNotFound，则 err 现在是 nil。
	// 如果原来就没有错误，则 err 保持 nil。
	// 我们现在只需要检查 existingUser 是否真的有值（在没有错误的情况下查到了用户）
	if existingUser != nil { // 如果 existingUser 不是 nil (并且 err 是 nil)，说明用户确实存在
		return nil, ErrUsernameExists // 用户名已存在。ErrUsernameExists 已定义在 service/user.go
	}

	// 如果执行到这里，说明用户名可用。
	// 2. 准备 User 模型
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		// Password 会在 DAO 层被哈希
		// UserUUID 会在 DAO 层生成
	}

	// 3. 创建用户 (DAO层负责哈希密码和生成UserUUID)
	if creationErr := mysql.CreateUser(user, req.Password); creationErr != nil {
		return nil, fmt.Errorf("创建用户失败: %w", creationErr)
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
		fmt.Printf("登录失败 - 密码验证失败: %v\n", err)
		return "", nil, ErrInvalidCredentials // Generic error for other bcrypt issues
	}

	// 3. 更新最后登录信息和在线状态
	if err := s.UpdateUserLastLoginInfo(user.ID); err != nil {
		fmt.Printf("警告: 更新用户最后登录信息失败: %v\n", err)
		// Non-critical error, proceed with login
	}
	if err := s.UpdateUserOnlineStatus(user.ID, true); err != nil {
		fmt.Printf("警告: 更新用户在线状态失败: %v\n", err)
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
		return "", nil, fmt.Errorf("生成JWT Token失败: %w", err)
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
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
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
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
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
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
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
