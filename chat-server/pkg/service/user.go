package service

import (
	"errors"
	"time"

	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrUserAlreadyExists = errors.New("用户已存在")
	ErrPasswordIncorrect = errors.New("密码错误")
)

// IUserService 用户服务接口, 提供用户注册、登录、获取用户信息等操作
type IUserService interface {
	// Register 注册新用户
	Register(req model.UserRegisterReq) (*model.User, error)

	// Login 用户登录
	Login(req model.UserLoginReq) (*model.User, error)

	// GetUserByID 根据用户ID获取用户信息
	GetUserByID(userID string) (*model.User, error)

	// GetUserByUsername 根据用户名获取用户信息
	GetUserByUsername(username string) (*model.User, error)

	// UpdateLastLogin 更新用户最后登录时间
	UpdateLastLogin(userID string) error
}

// InMemoryUserService 内存实现的用户服务
// TODO: 后续替换为数据库实现
type InMemoryUserService struct {
	users map[string]*model.User // userID -> User
}

// NewInMemoryUserService 创建内存用户服务
func NewInMemoryUserService() *InMemoryUserService {
	return &InMemoryUserService{
		users: make(map[string]*model.User),
	}
}

// Register 注册新用户
func (s *InMemoryUserService) Register(req model.UserRegisterReq) (*model.User, error) {
	// 检查用户是否已存在
	for _, existingUser := range s.users {
		if existingUser.Username == req.Username {
			return nil, ErrUserAlreadyExists
		}
	}

	// 生成用户ID
	userID := uuid.New().String()
	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &model.User{
		UserID:       userID,
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		Email:        req.Email,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LastLogin:    time.Now(),
	}

	// 保存用户
	s.users[userID] = user

	return user, nil
}

// Login 用户登录
func (s *InMemoryUserService) Login(req model.UserLoginReq) (*model.User, error) {
	// 查找用户
	var user *model.User
	for _, u := range s.users {
		if u.Username == req.Username {
			user = u
			break
		}
	}

	// 如果用户不存在
	if user == nil {
		return nil, ErrUserNotFound
	}

	// 验证密码
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, ErrPasswordIncorrect
	}

	// 更新最后登录时间
	s.UpdateLastLogin(user.UserID)

	return user, nil
}

// GetUserByID 根据用户ID获取用户信息
func (s *InMemoryUserService) GetUserByID(userID string) (*model.User, error) {
	user, ok := s.users[userID]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetUserByUsername 根据用户名获取用户信息
func (s *InMemoryUserService) GetUserByUsername(username string) (*model.User, error) {
	for _, user := range s.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

// UpdateLastLogin 更新用户最后登录时间
func (s *InMemoryUserService) UpdateLastLogin(userID string) error {
	user, ok := s.users[userID]
	if !ok {
		return ErrUserNotFound
	}

	user.LastLogin = time.Now()
	return nil
}
 