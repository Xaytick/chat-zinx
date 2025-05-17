package service

import (
	"errors"

	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("用户已存在")
	ErrPasswordIncorrect  = errors.New("password incorrect")
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid username or password")
)

// IUserService 定义用户服务接口
type IUserService interface {
	// Register 用户注册，成功返回完整的User模型（包含ID, UserUUID等）
	Register(req *model.UserRegisterReq) (*model.User, error)
	// Login 用户登录，成功返回JWT Token和User模型
	Login(req *model.UserLoginReq) (token string, user *model.User, err error)
	// GetUserByID 根据主键ID (uint) 获取用户信息
	GetUserByID(userID uint) (*model.User, error)
	// GetUserByUUID 根据UUID获取用户信息
	GetUserByUUID(uuid string) (*model.User, error)
	// GetUserByUsername 根据用户名获取用户信息
	GetUserByUsername(username string) (*model.User, error)
	// UpdateUserOnlineStatus 更新用户在线状态
	UpdateUserOnlineStatus(userID uint, isOnline bool) error
	// UpdateUserLastLoginInfo 更新用户最后登录信息
	UpdateUserLastLoginInfo(userID uint) error
}

/*
// InMemoryUserService 是IUserService的一个内存实现，主要用于测试
type InMemoryUserService struct {
	users map[string]*model.User // 使用 UserID 作为 key
	mu    sync.Mutex
}

// NewInMemoryUserService 创建一个新的内存用户服务实例
func NewInMemoryUserService() IUserService {
	return &InMemoryUserService{
		users: make(map[string]*model.User),
	}
}

// Register 注册新用户 (内存实现)
func (s *InMemoryUserService) Register(req *model.UserRegisterReq) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[req.Username]; exists { // 简单假设username是唯一的key
		return nil, errors.New("username already exists")
	}

	// 在内存实现中，我们可能不会处理真实的密码哈希和UserUUID生成，或者简化处理
	newUser := &model.User{
		// ID:        uint(len(s.users) + 1), // 简单的自增ID
		// UserUUID:  uuid.NewString(),
		Username: req.Username,
		// Password: req.Password, // 注意：不应该存储明文密码
		Email:    req.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		LastLogin: time.Now(),
	}
	s.users[newUser.Username] = newUser // 假设用Username作为key
	return newUser, nil
}

// Login 用户登录 (内存实现)
func (s *InMemoryUserService) Login(req *model.UserLoginReq) (string, *model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[req.Username]
	if !exists {
		return "", nil, errors.New("user not found")
	}

	// 内存实现中简化密码验证
	// if user.Password != req.Password { // 错误：不应该比较明文
	// 	 return "", nil, errors.New("invalid password")
	// }
	fmt.Println("Warning: In-memory login does not perform real password validation.")

 	// 模拟JWT token生成
	return "fake-jwt-token-for-in-memory", user, nil
}

// GetUserByID 根据用户ID获取用户信息 (内存实现)
func (s *InMemoryUserService) GetUserByID(userID uint) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, user := range s.users {
		// if user.ID == userID { // 假设 User 结构有 ID 字段
		// 	 return user, nil
		// }
	}
	fmt.Printf("Warning: In-memory GetUserByID (uint) not fully implemented to match ID field.\n")
	return nil, errors.New("user not found in-memory by uint ID")
}
*/
