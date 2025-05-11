package service

import (
	"github.com/Xaytick/chat-zinx/chat-server/pkg/storage"
)

// IMessageService 消息服务接口
type IMessageService interface {
	// SaveMessage 保存消息
	SaveMessage(fromUserID, toUserID string, msgData []byte) error

	// SaveHistoryOnly 只保存历史消息，不保存离线消息
	SaveHistoryOnly(fromUserID, toUserID string, msgData []byte) error

	// GetOfflineMessages 获取并清空用户的离线消息
	GetOfflineMessages(userID string) ([][]byte, error)

	// HasOfflineMessages 检查用户是否有离线消息
	HasOfflineMessages(userID string) bool

	// GetHistoryMessages 获取历史消息
	GetHistoryMessages(userID1, userID2 string, limit int) ([]map[string]interface{}, error)

	// GetChatRelations 获取用户的聊天关系列表
	GetChatRelations(userID string) ([]string, error)
}

// RedisMessageService Redis实现的消息服务
type RedisMessageService struct {
	storage *storage.RedisMsgStorage
}

// NewRedisMessageService 创建Redis消息服务
func NewRedisMessageService() *RedisMessageService {
	return &RedisMessageService{
		storage: storage.NewRedisMsgStorage(),
	}
}

// SaveMessage 保存消息
func (s *RedisMessageService) SaveMessage(fromUserID, toUserID string, msgData []byte) error {
	return s.storage.SaveMessage(fromUserID, toUserID, msgData)
}

// SaveHistoryOnly 只保存历史消息，不保存离线消息
func (s *RedisMessageService) SaveHistoryOnly(fromUserID, toUserID string, msgData []byte) error {
	return s.storage.SaveHistoryOnly(fromUserID, toUserID, msgData)
}

// GetOfflineMessages 获取并清空用户的离线消息
func (s *RedisMessageService) GetOfflineMessages(userID string) ([][]byte, error) {
	return s.storage.GetOfflineMessages(userID)
}

// HasOfflineMessages 检查用户是否有离线消息
func (s *RedisMessageService) HasOfflineMessages(userID string) bool {
	has, err := s.storage.HasOfflineMessages(userID)
	if err != nil {
		return false
	}
	return has
}

// GetHistoryMessages 获取历史消息
func (s *RedisMessageService) GetHistoryMessages(userID1, userID2 string, limit int) ([]map[string]interface{}, error) {
	return s.storage.GetHistoryMessages(userID1, userID2, int64(limit))
}

// GetChatRelations 获取用户的聊天关系列表
func (s *RedisMessageService) GetChatRelations(userID string) ([]string, error) {
	return s.storage.GetChatRelations(userID)
}
