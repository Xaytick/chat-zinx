package conf

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	MySQL MySQLConfig `json:"MySQL"` // MySQL配置
	Redis RedisConfig `json:"Redis"` // Redis配置
}

// MySQLConfig MySQL配置结构体
type MySQLConfig struct {
	Host         string `json:"Host"`         // 主机地址
	Port         int    `json:"Port"`         // 端口号
	User         string `json:"User"`         // 用户名
	Password     string `json:"Password"`     // 密码
	Database     string `json:"Database"`     // 数据库名称
	MaxOpenConns int    `json:"MaxOpenConns"` // 最大连接数
	MaxIdleConns int    `json:"MaxIdleConns"` // 最大空闲连接数
}

// RedisConfig Redis配置结构体
type RedisConfig struct {
	Host              string `json:"Host"`
	Port              int    `json:"Port"`
	Password          string `json:"Password"`
	DB                int    `json:"DB"`
	MessageExpiration int    `json:"MessageExpiration"`
}

// Config 应用配置结构体
type Config struct {
	Name           string         `json:"Name"`
	Host           string         `json:"Host"`
	TcpPort        int            `json:"TcpPort"`
	MaxConn        int            `json:"MaxConn"`
	WorkerPoolSize int            `json:"WorkerPoolSize"`
	MaxMsgChanLen  int            `json:"MaxMsgChanLen"`
	MaxPacketSize  int            `json:"MaxPacketSize"`
	Database       DatabaseConfig `json:"Database"`
}

// 全局配置实例
var GlobalConfig *Config

// LoadConfig 从文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	// 读取配置文件
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// 解析JSON配置
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// 更新全局配置
	GlobalConfig = &config
	return &config, nil
}

// GetMySQLConfig 获取MySQL配置
func GetMySQLConfig() *MySQLConfig {
	if GlobalConfig == nil {
		return nil
	}
	mysqlConfig := GlobalConfig.Database.MySQL
	return &mysqlConfig
}

// GetRedisConfig 获取Redis配置
func GetRedisConfig() *RedisConfig {
	if GlobalConfig == nil {
		return nil
	}
	redisConfig := GlobalConfig.Database.Redis
	return &redisConfig
}
