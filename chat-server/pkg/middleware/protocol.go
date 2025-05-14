package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
)

// AuthMessage 包含认证信息的消息结构
type AuthMessage struct {
	// 认证信息
	Auth struct {
		// 时间戳（Unix时间戳，秒）
		Timestamp string `json:"timestamp"`
		// 随机字符串，用于防重放攻击
		Nonce string `json:"nonce"`
		// 签名
		Signature string `json:"signature"`
		// JWT令牌
		Token string `json:"token"`
		// 用户ID，用于Redis会话验证
		UserID string `json:"user_id"`
	} `json:"auth"`

	// 其他消息内容，根据实际协议格式定义
	// 例如：业务数据
	Data interface{} `json:"data"`
}

// ExtractAuthFromJSON 从JSON格式的消息中提取认证信息
func ExtractAuthFromJSON(data []byte) (*AuthMessage, error) {
	var msg AuthMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("解析JSON消息失败: %v", err)
	}
	return &msg, nil
}

// ExportExtractSignatureParams 提取签名参数
func ExportExtractSignatureParams(data []byte) (timestamp string, nonce string, signature string, err error) {
	// 尝试从JSON中提取
	authMsg, err := ExtractAuthFromJSON(data)
	if err != nil {
		return "", "", "", err
	}

	// 验证参数
	if authMsg.Auth.Timestamp == "" {
		return "", "", "", errors.New("时间戳为空")
	}
	if authMsg.Auth.Nonce == "" {
		return "", "", "", errors.New("nonce为空")
	}
	if authMsg.Auth.Signature == "" {
		return "", "", "", errors.New("签名为空")
	}

	return authMsg.Auth.Timestamp, authMsg.Auth.Nonce, authMsg.Auth.Signature, nil
}

// ExportExtractToken 提取Token
func ExportExtractToken(data []byte) (string, error) {
	// 尝试从JSON中提取
	authMsg, err := ExtractAuthFromJSON(data)
	if err != nil {
		return "", err
	}

	// 验证参数
	if authMsg.Auth.Token == "" {
		return "", errors.New("token为空")
	}

	return authMsg.Auth.Token, nil
}

// ExportExtractUserIDAndToken 提取用户ID和Token
func ExportExtractUserIDAndToken(data []byte) (string, string, error) {
	// 尝试从JSON中提取
	authMsg, err := ExtractAuthFromJSON(data)
	if err != nil {
		return "", "", err
	}

	// 验证参数
	if authMsg.Auth.UserID == "" {
		return "", "", errors.New("用户ID为空")
	}
	if authMsg.Auth.Token == "" {
		return "", "", errors.New("token为空")
	}

	return authMsg.Auth.UserID, authMsg.Auth.Token, nil
}

// GetAuthMessageData 从认证消息中获取实际的业务数据
func GetAuthMessageData(data []byte) ([]byte, error) {
	// 解析认证消息
	authMsg, err := ExtractAuthFromJSON(data)
	if err != nil {
		return nil, err
	}

	// 提取业务数据部分
	if authMsg.Data == nil {
		return nil, errors.New("消息中不包含业务数据")
	}

	// 将业务数据序列化为JSON
	return json.Marshal(authMsg.Data)
}
