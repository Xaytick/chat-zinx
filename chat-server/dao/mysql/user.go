package mysql

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Xaytick/chat-zinx/chat-server/conf"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"github.com/google/uuid" // 用于生成 UserUUID
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var ErrRecordNotFound = gorm.ErrRecordNotFound // Export gorm.ErrRecordNotFound

// InitMySQL 初始化MySQL数据库连接 (使用GORM)
func InitMySQL(cfg *conf.MySQLConfig) (err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: false, // Set to false to handle it explicitly
			Colorful:                  true,
		},
	)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 自动迁移时，请确保您的 User 模型与数据库表结构匹配 GORM 的约定或使用了正确的 gorm tags
	err = DB.AutoMigrate(&model.User{}, &model.Group{}, &model.GroupMember{}) // Group 和 GroupMember 已是 ID uint 主键
	if err != nil {
		return fmt.Errorf("failed to auto migrate tables: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get generic database object from GORM: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	// sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute) // ConnMaxLifetime 在当前 MySQLConfig 中未定义

	log.Println("GORM MySQL connection pool initialized successfully.")
	return nil
}

// GetUserByUsername 根据用户名查询用户信息 (GORM实现)
func GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	result := DB.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound // Return exported error
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByID 根据主键ID查询用户信息 (GORM实现)
func GetUserByID(id uint) (*model.User, error) {
	var user model.User
	result := DB.First(&user, id) // GORM 默认按主键查找
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound // Return exported error
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByUUID 根据UserUUID查询用户信息 (GORM实现)
func GetUserByUUID(uuidString string) (*model.User, error) {
	var user model.User
	result := DB.Where("user_uuid = ?", uuidString).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound // Return exported error
		}
		return nil, result.Error
	}
	return &user, nil
}

// CreateUser 创建用户 (GORM实现)
// plainPassword 是明文密码，此函数负责哈希
func CreateUser(user *model.User, plainPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword)

	// 生成 UserUUID
	user.UserUUID = uuid.NewString()

	// GORM autoCreateTime handles CreatedAt
	// GORM autoUpdateTime handles UpdatedAt upon creation as well
	// IsOnline defaults to false in the model struct tag
	// LastLogin can be set to current time on creation or be nil/zero
	user.LastLogin = time.Now()

	result := DB.Create(user)
	return result.Error
}

// UpdateUserOnlineStatus 更新用户在线状态 (GORM实现)
func UpdateUserOnlineStatus(userID uint, isOnline bool) error {
	updates := map[string]interface{}{
		"is_online":  isOnline,
		"updated_at": time.Now(),
	}
	return DB.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

// UpdateUserLastLoginInfo 更新用户最后登录时间 (GORM实现)
func UpdateUserLastLoginInfo(userID uint) error {
	updates := map[string]interface{}{
		"last_login": time.Now(),
		"updated_at": time.Now(),
	}
	return DB.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}
