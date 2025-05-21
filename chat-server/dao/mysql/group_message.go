package mysql

import (
	"fmt"
	"time"

	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"github.com/google/uuid"
)

// SaveGroupMessage 保存群组消息到数据库
func SaveGroupMessage(groupID uint, senderID uint, senderUUID, senderName, content string, messageType string) (*model.GroupMessage, error) {
	message := &model.GroupMessage{
		MsgID:       uuid.NewString(), // 生成消息唯一ID
		GroupID:     groupID,
		SenderID:    senderID,
		SenderUUID:  senderUUID,
		SenderName:  senderName,
		Content:     content,
		MessageType: messageType,
		CreatedAt:   time.Now(),
	}

	if err := DB.Create(message).Error; err != nil {
		return nil, fmt.Errorf("failed to save group message: %w", err)
	}

	return message, nil
}

// GetGroupHistoryMessages 获取群组历史消息
func GetGroupHistoryMessages(groupID uint, lastID uint, limit int) ([]*model.GroupMessage, bool, error) {
	if limit <= 0 {
		limit = 20 // 默认限制为20条
	}
	if limit > 100 {
		limit = 100 // 最大限制为100条
	}

	var messages []*model.GroupMessage
	query := DB.Where("group_id = ?", groupID)

	if lastID > 0 {
		// 如果提供了lastID，获取比这个ID小的消息（更早的消息）
		query = query.Where("id < ?", lastID)
	}

	// 按ID降序排列，获取最新的消息
	err := query.Order("id DESC").Limit(limit + 1).Find(&messages).Error
	if err != nil {
		return nil, false, fmt.Errorf("failed to get group history messages: %w", err)
	}

	// 检查是否还有更多消息
	hasMore := false
	if len(messages) > limit {
		hasMore = true
		messages = messages[:limit] // 切掉多查询的一条
	}

	return messages, hasMore, nil
}
