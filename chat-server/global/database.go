package global

import (
	"encoding/json"
	"log"
	"sync"

	"os"

	"github.com/Xaytick/chat-zinx/chat-server/pkg/database"
)

var (
	// DatabaseManager 全局数据库管理器
	DatabaseManager *database.DatabaseManager

	// Repository 全局数据访问层
	Repository *database.Repository

	// dbOnce 确保数据库只初始化一次
	dbOnce sync.Once
)

// InitDatabase 初始化数据库连接
func InitDatabase(configPath string) error {
	var err error

	dbOnce.Do(func() {
		// 读取配置文件
		configData, readErr := os.ReadFile(configPath)
		if readErr != nil {
			err = readErr
			return
		}

		// 解析JSON配置
		var config struct {
			Database database.DatabaseConfig `json:"database"`
		}

		if parseErr := json.Unmarshal(configData, &config); parseErr != nil {
			err = parseErr
			return
		}

		// 创建数据库管理器
		DatabaseManager, err = database.NewDatabaseManager(&config.Database)
		if err != nil {
			return
		}

		// 创建数据访问层
		Repository = database.NewRepository(DatabaseManager)

		// 执行健康检查
		if healthErr := DatabaseManager.Health(); healthErr != nil {
			log.Printf("Database health check warning: %v", healthErr)
		}

		log.Println("Database initialized successfully")
	})

	return err
}

// CloseDatabase 关闭数据库连接
func CloseDatabase() error {
	if DatabaseManager != nil {
		return DatabaseManager.Close()
	}
	return nil
}

// GetDatabase 获取数据库管理器
func GetDatabase() *database.DatabaseManager {
	return DatabaseManager
}

// GetRepository 获取数据访问层
func GetRepository() *database.Repository {
	return Repository
}

// InitDatabaseWithJSON 使用JSON字符串初始化数据库（用于测试）
func InitDatabaseWithJSON(jsonConfig string) error {
	var err error

	dbOnce.Do(func() {
		// 解析JSON配置
		var config struct {
			Database database.DatabaseConfig `json:"database"`
		}

		if parseErr := json.Unmarshal([]byte(jsonConfig), &config); parseErr != nil {
			err = parseErr
			return
		}

		// 创建数据库管理器
		DatabaseManager, err = database.NewDatabaseManager(&config.Database)
		if err != nil {
			return
		}

		// 创建数据访问层
		Repository = database.NewRepository(DatabaseManager)

		// 执行健康检查
		if healthErr := DatabaseManager.Health(); healthErr != nil {
			log.Printf("Database health check warning: %v", healthErr)
		}

		log.Println("Database initialized successfully with JSON config")
	})

	return err
}
