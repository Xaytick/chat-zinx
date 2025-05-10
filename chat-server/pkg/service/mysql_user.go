package service

import (
	"time"

	"github.com/Xaytick/chat-zinx/chat-server/dao/mysql"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// MySQLUserService MySQL实现的用户服务
type MySQLUserService struct{}

// NewMySQLUserService 创建MySQL用户服务
func NewMySQLUserService() *MySQLUserService {
	return &MySQLUserService{}
}

// Register 注册新用户
func (s *MySQLUserService) Register(req model.UserRegisterReq) (*model.User, error) {
	// 检查用户是否已存在
	_, err := mysql.GetUserByUsername(req.Username)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}

	// 生成用户ID
	userID := uuid.New().String()

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 当前时间
	now := time.Now()

	// 创建用户
	dbUser := &mysql.User{
		UserID:       userID,
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		Email:        req.Email,
		CreatedAt:    now,
		UpdatedAt:    now,
		LastLogin:    now,
	}

	// 保存用户到数据库
	_, err = mysql.InsertUser(dbUser)
	if err != nil {
		return nil, err
	}

	// 返回用户模型
	user := &model.User{
		UserID:       userID,
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		Email:        req.Email,
		CreatedAt:    now,
		UpdatedAt:    now,
		LastLogin:    now,
	}

	return user, nil
}

// Login 用户登录
func (s *MySQLUserService) Login(req model.UserLoginReq) (*model.User, error) {
	// 查找用户
	dbUser, err := mysql.GetUserByUsername(req.Username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, ErrPasswordIncorrect
	}

	// 更新最后登录时间
	err = mysql.UpdateLastLogin(dbUser.UserID)
	if err != nil {
		return nil, err
	}

	// 转换为模型
	user := &model.User{
		UserID:       dbUser.UserID,
		Username:     dbUser.Username,
		PasswordHash: dbUser.PasswordHash,
		Salt:         dbUser.Salt,
		Email:        dbUser.Email,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		LastLogin:    dbUser.LastLogin,
	}

	return user, nil
}

// GetUserByID 根据用户ID获取用户信息
func (s *MySQLUserService) GetUserByID(userID string) (*model.User, error) {
	dbUser, err := mysql.GetUserByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// 转换为模型
	user := &model.User{
		UserID:       dbUser.UserID,
		Username:     dbUser.Username,
		PasswordHash: dbUser.PasswordHash,
		Salt:         dbUser.Salt,
		Email:        dbUser.Email,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		LastLogin:    dbUser.LastLogin,
	}

	return user, nil
}

// GetUserByUsername 根据用户名获取用户信息
func (s *MySQLUserService) GetUserByUsername(username string) (*model.User, error) {
	dbUser, err := mysql.GetUserByUsername(username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// 转换为模型
	user := &model.User{
		UserID:       dbUser.UserID,
		Username:     dbUser.Username,
		PasswordHash: dbUser.PasswordHash,
		Salt:         dbUser.Salt,
		Email:        dbUser.Email,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		LastLogin:    dbUser.LastLogin,
	}

	return user, nil
}

// UpdateLastLogin 更新用户最后登录时间
func (s *MySQLUserService) UpdateLastLogin(userID string) error {
	return mysql.UpdateLastLogin(userID)
}
