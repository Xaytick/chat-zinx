package service

import (
	"fmt"

	"github.com/Xaytick/chat-zinx/chat-server/dao/mysql"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/storage"
)

// IMessageService 消息服务接口
type IMessageService interface {
	// SaveMessage 保存消息
	SaveMessage(fromUserID, toUserID uint, msgData []byte) error

	// SaveHistoryOnly 只保存历史消息，不保存离线消息
	SaveHistoryOnly(fromUserID, toUserID uint, msgData []byte) error

	// GetOfflineMessages 获取并清空用户的离线消息
	GetOfflineMessages(userID uint) ([][]byte, error)

	// HasOfflineMessages 检查用户是否有离线消息
	HasOfflineMessages(userID uint) bool

	// GetHistoryMessages 获取历史消息
	GetHistoryMessages(userID1, userID2 uint, limit int) ([]map[string]interface{}, error)

	// GetChatRelations 获取用户的聊天关系列表
	GetChatRelations(userID uint) ([]string, error)

	// P2P 消息相关
	SaveSingleMessage(fromUserID, toUserID uint, msgType, content string) error
	GetChatHistory(userID, peerUserID uint, lastMsgID string, limit int) ([]*model.MessageItemResp, error)

	// 群组消息相关
	SaveGroupMessage(groupID uint, senderID uint, senderUUID, senderName, content, messageType string) (string, error)
	GetGroupHistory(userID, groupID uint, lastID uint, limit int) (*model.GroupHistoryMsgResp, error)
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
func (s *RedisMessageService) SaveMessage(fromUserID, toUserID uint, msgData []byte) error {
	return s.storage.SaveMessage(fromUserID, toUserID, msgData)
}

// SaveHistoryOnly 只保存历史消息，不保存离线消息
func (s *RedisMessageService) SaveHistoryOnly(fromUserID, toUserID uint, msgData []byte) error {
	return s.storage.SaveHistoryOnly(fromUserID, toUserID, msgData)
}

// GetOfflineMessages 获取并清空用户的离线消息
func (s *RedisMessageService) GetOfflineMessages(userID uint) ([][]byte, error) {
	return s.storage.GetOfflineMessages(userID)
}

// HasOfflineMessages 检查用户是否有离线消息
func (s *RedisMessageService) HasOfflineMessages(userID uint) bool {
	has, err := s.storage.HasOfflineMessages(userID)
	if err != nil {
		return false
	}
	return has
}

// GetHistoryMessages 获取历史消息
func (s *RedisMessageService) GetHistoryMessages(userID1, userID2 uint, limit int) ([]map[string]interface{}, error) {
	return s.storage.GetHistoryMessages(userID1, userID2, int64(limit))
}

// GetChatRelations 获取用户的聊天关系列表
func (s *RedisMessageService) GetChatRelations(userID uint) ([]string, error) {
	return s.storage.GetChatRelations(userID)
}

// SaveSingleMessage 保存单聊消息
func (s *RedisMessageService) SaveSingleMessage(fromUserID, toUserID uint, msgType, content string) error {
	// 待实现
	return nil
}

// GetChatHistory 获取单聊历史消息
func (s *RedisMessageService) GetChatHistory(userID, peerUserID uint, lastMsgID string, limit int) ([]*model.MessageItemResp, error) {
	// 待实现
	return nil, nil
}

// SaveGroupMessage 保存群组消息
func (s *RedisMessageService) SaveGroupMessage(groupID uint, senderID uint, senderUUID, senderName, content, messageType string) (string, error) {
	// 使用MySQL保存群组消息
	if messageType == "" {
		messageType = "text" // 默认为文本消息
	}

	message, err := mysql.SaveGroupMessage(groupID, senderID, senderUUID, senderName, content, messageType)
	if err != nil {
		return "", fmt.Errorf("failed to save group message: %w", err)
	}

	return message.MsgID, nil
}

// GetGroupHistory 获取群组历史消息
func (s *RedisMessageService) GetGroupHistory(userID, groupID uint, lastID uint, limit int) (*model.GroupHistoryMsgResp, error) {
	// 1. 检查用户是否是群成员
	isMember, err := mysql.IsUserInGroup(userID, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to check group membership: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("user %d is not a member of group %d", userID, groupID)
	}

	// 2. 获取历史消息
	messages, hasMore, err := mysql.GetGroupHistoryMessages(groupID, lastID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get group history messages: %w", err)
	}

	// 3. 转换为响应格式
	msgItems := make([]*model.GroupHistoryMsgItem, 0, len(messages))
	for _, msg := range messages {
		msgItem := &model.GroupHistoryMsgItem{
			ID:          msg.ID,
			MsgID:       msg.MsgID,
			SenderID:    msg.SenderID,
			SenderUUID:  msg.SenderUUID,
			SenderName:  msg.SenderName,
			Content:     msg.Content,
			MessageType: msg.MessageType,
			Timestamp:   msg.CreatedAt.Unix(),
		}
		msgItems = append(msgItems, msgItem)
	}

	// 4. 构建并返回响应
	resp := &model.GroupHistoryMsgResp{
		GroupID:  groupID,
		Messages: msgItems,
		HasMore:  hasMore,
	}

	return resp, nil
}
