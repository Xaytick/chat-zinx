package examples

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"

	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
)

// 生成带认证信息的消息
func GenerateAuthMessage(msgData interface{}, token string) ([]byte, error) {
	// 创建认证消息结构
	authMsg := struct {
		Auth struct {
			Timestamp string `json:"timestamp"`
			Nonce     string `json:"nonce"`
			Signature string `json:"signature"`
			Token     string `json:"token"`
			UserID    string `json:"user_id"`
		} `json:"auth"`
		Data interface{} `json:"data"`
	}{}

	// 设置认证信息
	authMsg.Auth.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	authMsg.Auth.Nonce = generateNonce()
	authMsg.Auth.Token = token
	authMsg.Auth.UserID = extractUserIDFromToken(token) // 从Token中提取用户ID

	// 计算签名
	secretKey := "your-signature-secret-please-change-in-production" // 应从配置中获取
	authMsg.Auth.Signature = generateSignature(authMsg.Auth.Timestamp, authMsg.Auth.Nonce, secretKey)

	// 设置业务数据
	authMsg.Data = msgData

	// 序列化为JSON
	return json.Marshal(authMsg)
}

// 生成随机字符串作为nonce
func generateNonce() string {
	// 简单实现：时间戳+随机数
	return fmt.Sprintf("%d%d", time.Now().UnixNano(), time.Now().Nanosecond())
}

// 从Token中提取用户ID
func extractUserIDFromToken(token string) string {
	// 在实际应用中，应该解析JWT获取用户ID
	// 这里简化处理，假设已知用户ID
	return "user-id-123"
}

// 生成HMAC-SHA256签名
func generateSignature(timestamp, nonce, secretKey string) string {
	// 按字典序排序参数
	params := []string{timestamp, nonce, secretKey}
	sort.Strings(params)
	str := strings.Join(params, "")

	// 计算HMAC-SHA256签名
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// 示例：创建带认证信息的文本消息
func CreateAuthTextMessage(token string, toUserID, content string) ([]byte, error) {
	// 创建文本消息
	textMsg := model.TextMsg{
		ToUserID: toUserID,
		Content:  content,
	}

	// 添加认证信息
	return GenerateAuthMessage(textMsg, token)
}

// 示例：创建带认证信息的历史消息请求
func CreateAuthHistoryRequest(token string, targetUserID string, limit int) ([]byte, error) {
	// 创建历史消息请求
	historyReq := struct {
		TargetUserID string `json:"target_user_id"`
		Limit        int    `json:"limit"`
	}{
		TargetUserID: targetUserID,
		Limit:        limit,
	}

	// 添加认证信息
	return GenerateAuthMessage(historyReq, token)
}
