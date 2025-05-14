package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Xaytick/chat-zinx/chat-server/conf"
	"github.com/Xaytick/chat-zinx/chat-server/dao/redis"
	"github.com/Xaytick/chat-zinx/chat-server/global"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"github.com/Xaytick/zinx/ziface"
	"github.com/golang-jwt/jwt"
)

// 自定义JWT声明
type CustomClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.StandardClaims
}

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	// 配置选项
	EnableSignatureCheck bool   // 是否启用签名校验
	EnableJWTCheck       bool   // 是否启用JWT校验
	EnableRedisCheck     bool   // 是否启用Redis会话验证
	SecretKey            string // 签名密钥
	JWTSecret            string // JWT密钥
	TimestampTolerance   int64  // 时间戳容忍误差（秒）
	NonceExpiration      int64  // nonce在Redis中的过期时间（秒）
	SessionExpiration    int64  // 会话过期时间（秒）
}

// NewAuthMiddleware 创建新的认证中间件
func NewAuthMiddleware(options ...func(*AuthMiddleware)) *AuthMiddleware {
	// 从配置中获取默认设置
	authConfig := conf.GetAuthConfig()

	// 默认配置
	m := &AuthMiddleware{
		EnableSignatureCheck: true,
		EnableJWTCheck:       true,
		EnableRedisCheck:     false, // 默认不启用Redis校验，减少性能开销
		SecretKey:            authConfig.SignatureSecret,
		JWTSecret:            authConfig.JWT.Secret,
		TimestampTolerance:   int64(authConfig.Security.TimestampTolerance),
		NonceExpiration:      int64(authConfig.Security.NonceExpiration),
		SessionExpiration:    int64(authConfig.Security.SessionExpiration),
	}

	// 应用自定义配置
	for _, opt := range options {
		opt(m)
	}

	return m
}

// WithSignatureCheck 设置是否启用签名校验
func WithSignatureCheck(enable bool) func(*AuthMiddleware) {
	return func(m *AuthMiddleware) {
		m.EnableSignatureCheck = enable
	}
}

// WithJWTCheck 设置是否启用JWT校验
func WithJWTCheck(enable bool) func(*AuthMiddleware) {
	return func(m *AuthMiddleware) {
		m.EnableJWTCheck = enable
	}
}

// WithRedisCheck 设置是否启用Redis会话校验
func WithRedisCheck(enable bool) func(*AuthMiddleware) {
	return func(m *AuthMiddleware) {
		m.EnableRedisCheck = enable
	}
}

// WithSecretKey 设置签名密钥
func WithSecretKey(key string) func(*AuthMiddleware) {
	return func(m *AuthMiddleware) {
		m.SecretKey = key
	}
}

// WithJWTSecret 设置JWT密钥
func WithJWTSecret(key string) func(*AuthMiddleware) {
	return func(m *AuthMiddleware) {
		m.JWTSecret = key
	}
}

// Verify 验证请求
// 返回：认证结果（通过/失败），用户信息（如果认证通过），错误信息
func (m *AuthMiddleware) Verify(request ziface.IRequest) (bool, *model.User, error) {
	// 从请求中提取认证信息
	// 注意：这里假设认证信息是通过请求头或请求参数传递的
	// 实际实现中需要根据具体协议格式进行提取

	// 示例：从消息体中提取认证信息
	// 这里需要根据你的协议格式进行调整
	authData := request.GetData()
	if len(authData) == 0 {
		return false, nil, errors.New("认证数据为空")
	}

	// 1. 验证签名（如果启用）
	if m.EnableSignatureCheck {
		timestamp, nonce, signature, err := ExportExtractSignatureParams(authData)
		if err != nil {
			return false, nil, fmt.Errorf("提取签名参数失败: %v", err)
		}

		if !m.verifySignature(timestamp, nonce, signature) {
			return false, nil, errors.New("签名验证失败")
		}
	}

	// 2. 验证JWT（如果启用）
	var userInfo *model.User
	var jwtErr error

	if m.EnableJWTCheck {
		token, err := ExportExtractToken(authData)
		if err != nil {
			return false, nil, fmt.Errorf("提取Token失败: %v", err)
		}

		claims, err := m.verifyJWT(token)
		if err != nil {
			jwtErr = err
			// 继续处理，如果Redis会话验证启用的话
		} else {
			// JWT验证通过，获取用户信息
			userInfo, err = global.UserService.GetUserByID(claims.UserID)
			if err != nil {
				return false, nil, fmt.Errorf("获取用户信息失败: %v", err)
			}

			// 3. 验证Redis会话（如果启用）
			if m.EnableRedisCheck {
				if !m.verifyRedisSession(claims.UserID, token) {
					return false, nil, errors.New("会话验证失败")
				}
			}

			return true, userInfo, nil
		}
	}

	// 如果JWT验证失败但启用了Redis会话验证
	if m.EnableRedisCheck && jwtErr != nil {
		// 从请求中提取用户ID和Token
		userID, token, err := ExportExtractUserIDAndToken(authData)
		if err != nil {
			return false, nil, fmt.Errorf("提取用户ID和Token失败: %v", err)
		}

		// 验证Redis会话
		if !m.verifyRedisSession(userID, token) {
			return false, nil, errors.New("会话验证失败")
		}

		// Redis验证通过，获取用户信息
		userInfo, err = global.UserService.GetUserByID(userID)
		if err != nil {
			return false, nil, fmt.Errorf("获取用户信息失败: %v", err)
		}

		return true, userInfo, nil
	}

	// 如果所有验证都失败
	return false, nil, errors.New("认证失败")
}

// verifySignature 验证签名
func (m *AuthMiddleware) verifySignature(timestamp string, nonce string, signature string) bool {
	// 1. 检查时间戳是否在有效期内
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		fmt.Printf("[Auth] 时间戳格式错误: %v\n", err)
		return false
	}

	now := time.Now().Unix()
	if now-ts > m.TimestampTolerance || ts-now > m.TimestampTolerance {
		fmt.Printf("[Auth] 时间戳超出容忍范围: 当前=%d, 请求=%d\n", now, ts)
		return false
	}

	// 2. 检查nonce是否已使用
	nonceKey := fmt.Sprintf("nonce:%s", nonce)
	exists, err := redis.RedisClient.Exists(redis.Ctx, nonceKey).Result()
	if err != nil {
		fmt.Printf("[Auth] 检查nonce失败: %v\n", err)
		return false
	}
	if exists == 1 {
		fmt.Printf("[Auth] nonce已使用: %s\n", nonce)
		return false
	}

	// 3. 验证签名
	params := []string{timestamp, nonce, m.SecretKey}
	sort.Strings(params)
	str := strings.Join(params, "")

	h := hmac.New(sha256.New, []byte(m.SecretKey))
	h.Write([]byte(str))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	if signature != expectedSignature {
		fmt.Printf("[Auth] 签名不匹配: 期望=%s, 实际=%s\n", expectedSignature, signature)
		return false
	}

	// 4. 记录已使用的nonce
	err = redis.RedisClient.Set(redis.Ctx, nonceKey, "1", time.Duration(m.NonceExpiration)*time.Second).Err()
	if err != nil {
		fmt.Printf("[Auth] 记录nonce失败: %v\n", err)
		// 不影响验证结果
	}

	return true
}

// verifyJWT 验证JWT
func (m *AuthMiddleware) verifyJWT(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的token")
}

// verifyRedisSession 验证Redis会话
func (m *AuthMiddleware) verifyRedisSession(userID string, token string) bool {
	sessionKey := fmt.Sprintf("session:%s", userID)
	storedToken, err := redis.RedisClient.Get(redis.Ctx, sessionKey).Result()
	if err != nil {
		fmt.Printf("[Auth] 获取会话失败: %v\n", err)
		return false
	}

	return token == storedToken
}

// GenerateToken 生成JWT令牌
func (m *AuthMiddleware) GenerateToken(userID, username string) (string, error) {
	// 创建Claims
	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(m.SessionExpiration) * time.Second).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    conf.GetAuthConfig().JWT.Issuer,
		},
	}

	// 创建JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名并获取完整的编码后的字符串令牌
	return token.SignedString([]byte(m.JWTSecret))
}

// SaveSession 保存会话到Redis
func (m *AuthMiddleware) SaveSession(userID, token string) error {
	sessionKey := fmt.Sprintf("session:%s", userID)
	return redis.RedisClient.Set(redis.Ctx, sessionKey, token, time.Duration(m.SessionExpiration)*time.Second).Err()
}

// RemoveSession 从Redis移除会话
func (m *AuthMiddleware) RemoveSession(userID string) error {
	sessionKey := fmt.Sprintf("session:%s", userID)
	return redis.RedisClient.Del(redis.Ctx, sessionKey).Err()
}

// 这些函数仅保留声明，实际实现调用protocol.go中的函数
// 避免重复声明错误
