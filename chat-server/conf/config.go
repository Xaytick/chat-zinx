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
	Host              string `json:"Host"`              // 主机地址
	Port              int    `json:"Port"`              // 端口号
	Password          string `json:"Password"`          // 密码
	DB                int    `json:"DB"`                // 数据库
	MessageExpiration int    `json:"MessageExpiration"` // 消息过期时间

	// Redis Cluster 配置
	ClusterEnabled bool     `json:"ClusterEnabled"` // 是否启用集群模式
	ClusterAddrs   []string `json:"ClusterAddrs"`   // 集群节点地址
	PoolSize       int      `json:"PoolSize"`       // 连接池大小
	MinIdleConns   int      `json:"MinIdleConns"`   // 最小空闲连接数
	MaxRetries     int      `json:"MaxRetries"`     // 最大重试次数
}

// HeartbeatConfig 心跳配置结构体
type HeartbeatConfig struct {
	Enabled  bool `json:"Enabled"`  // 是否启用心跳检测
	Interval int  `json:"Interval"` // 心跳检测间隔（秒）
	Timeout  int  `json:"Timeout"`  // 心跳超时时间（秒）
}

// AuthConfig 认证配置结构体
type AuthConfig struct {
	JWT             JWTConfig      `json:"JWT"`             // JWT配置
	Security        SecurityConfig `json:"Security"`        // 安全配置
	SignatureSecret string         `json:"SignatureSecret"` // 签名密钥
}

// JWTConfig JWT配置结构体
type JWTConfig struct {
	Secret    string `json:"Secret"`    // 密钥
	ExpiresIn int    `json:"ExpiresIn"` // 过期时间
	Issuer    string `json:"Issuer"`    // 发行者
}

// SecurityConfig 安全配置结构体
type SecurityConfig struct {
	TimestampTolerance int `json:"TimestampTolerance"` // 时间戳容忍度
	NonceExpiration    int `json:"NonceExpiration"`    // 非对称加密过期时间
	SessionExpiration  int `json:"SessionExpiration"`  // 会话过期时间
}

// Config 应用配置结构体
type Config struct {
	Name           string          `json:"Name"`           // 名称
	Host           string          `json:"Host"`           // 主机地址
	TcpPort        int             `json:"TcpPort"`        // 端口号
	MaxConn        int             `json:"MaxConn"`        // 最大连接数
	WorkerPoolSize int             `json:"WorkerPoolSize"` // 工作池大小
	MaxMsgChanLen  int             `json:"MaxMsgChanLen"`  // 最大消息通道长度
	MaxPacketSize  int             `json:"MaxPacketSize"`  // 最大包大小
	Heartbeat      HeartbeatConfig `json:"Heartbeat"`      // 心跳配置
	Database       DatabaseConfig  `json:"Database"`       // 数据库配置
	Auth           AuthConfig      `json:"Auth"`           // 认证配置
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

	// 设置默认值
	setDefaultAuthConfig(&config.Auth)
	setDefaultHeartbeatConfig(&config.Heartbeat)

	// 更新全局配置
	GlobalConfig = &config
	return &config, nil
}

// 设置认证配置默认值
func setDefaultAuthConfig(authConfig *AuthConfig) {
	// JWT配置默认值
	if authConfig.JWT.Secret == "" {
		authConfig.JWT.Secret = "default-jwt-secret-please-change-in-production"
	}
	if authConfig.JWT.ExpiresIn == 0 {
		authConfig.JWT.ExpiresIn = 86400 // 24小时
	}
	if authConfig.JWT.Issuer == "" {
		authConfig.JWT.Issuer = "chat-zinx"
	}

	// 安全配置默认值
	if authConfig.Security.TimestampTolerance == 0 {
		authConfig.Security.TimestampTolerance = 300 // 5分钟
	}
	if authConfig.Security.NonceExpiration == 0 {
		authConfig.Security.NonceExpiration = 600 // 10分钟
	}
	if authConfig.Security.SessionExpiration == 0 {
		authConfig.Security.SessionExpiration = 86400 // 24小时
	}

	// 签名密钥默认值
	if authConfig.SignatureSecret == "" {
		authConfig.SignatureSecret = "default-signature-secret-please-change-in-production"
	}
}

// 设置心跳配置默认值
func setDefaultHeartbeatConfig(heartbeatConfig *HeartbeatConfig) {
	// 如果没有设置，默认启用心跳
	if heartbeatConfig.Interval == 0 {
		heartbeatConfig.Interval = 60 // 默认60秒
	}
	if heartbeatConfig.Timeout == 0 {
		heartbeatConfig.Timeout = 180 // 默认180秒
	}
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

// GetAuthConfig 获取认证配置
func GetAuthConfig() *AuthConfig {
	if GlobalConfig == nil {
		return nil
	}
	authConfig := GlobalConfig.Auth
	return &authConfig
}

// GetHeartbeatConfig 获取心跳配置
func GetHeartbeatConfig() *HeartbeatConfig {
	if GlobalConfig == nil {
		return nil
	}
	heartbeatConfig := GlobalConfig.Heartbeat
	return &heartbeatConfig
}

// GetHeartbeatInterval 获取心跳间隔时间
func GetHeartbeatInterval() int {
	config := GetHeartbeatConfig()
	if config == nil {
		return 60 // 默认60秒
	}
	return config.Interval
}

// GetHeartbeatTimeout 获取心跳超时时间
func GetHeartbeatTimeout() int {
	config := GetHeartbeatConfig()
	if config == nil {
		return 180 // 默认180秒
	}
	return config.Timeout
}

// IsHeartbeatEnabled 检查心跳是否启用
func IsHeartbeatEnabled() bool {
	config := GetHeartbeatConfig()
	if config == nil {
		return true // 默认启用
	}
	return config.Enabled
}
