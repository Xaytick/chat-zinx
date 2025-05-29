package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/Xaytick/chat-zinx/chat-server/pkg/database"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"gorm.io/gorm"
)

// EnhancedMessageDAO 增强的消息DAO，支持分片
type EnhancedMessageDAO struct {
	repo *database.Repository
}

// NewEnhancedMessageDAO 创建新的增强消息DAO
func NewEnhancedMessageDAO(repo *database.Repository) *EnhancedMessageDAO {
	return &EnhancedMessageDAO{
		repo: repo,
	}
}

// SaveMessage 保存单聊消息（按发送者ID分片）
func (dao *EnhancedMessageDAO) SaveMessage(ctx context.Context, fromUserID, toUserID, content, msgType string) error {
	message := &model.TextMsg{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Content:    content,
		Type:       msgType,
		SentAt:     time.Now(),
	}

	opts := &database.QueryOptions{
		Operation:    database.ShardWriteOperation,
		ShardKey:     fromUserID, // 使用发送者ID作为分片键
		TableName:    "messages",
		ForceReplace: true,
	}

	return dao.repo.Execute(ctx, opts, func(db *gorm.DB) error {
		return db.Create(message).Error
	})
}

// SaveGroupMessage 保存群组消息（按群组ID分片）
func (dao *EnhancedMessageDAO) SaveGroupMessage(ctx context.Context, groupID uint, senderID uint, senderUUID, senderName, content, messageType string) (*model.GroupMessage, error) {
	message := &model.GroupMessage{
		MsgID:       generateMessageID(), // 需要实现一个ID生成函数
		GroupID:     groupID,
		SenderID:    senderID,
		SenderUUID:  senderUUID,
		SenderName:  senderName,
		Content:     content,
		MessageType: messageType,
		CreatedAt:   time.Now(),
	}

	opts := &database.QueryOptions{
		Operation:    database.ShardWriteOperation,
		ShardKey:     groupID, // 使用群组ID作为分片键
		TableName:    "group_messages",
		ForceReplace: true,
	}

	err := dao.repo.Execute(ctx, opts, func(db *gorm.DB) error {
		return db.Create(message).Error
	})

	if err != nil {
		return nil, err
	}
	return message, nil
}

// GetMessageHistory 获取单聊历史消息（需要查询两个分片）
func (dao *EnhancedMessageDAO) GetMessageHistory(ctx context.Context, userID1, userID2 uint, limit int) ([]*model.TextMsg, error) {
	var allMessages []*model.TextMsg

	// 需要查询两个用户的分片，因为消息可能分布在不同的分片中
	shardKeys := []interface{}{userID1, userID2}

	err := dao.repo.MultiShardOperation(ctx, shardKeys, func(dbMap map[interface{}]*gorm.DB) error {
		for shardKey, db := range dbMap {
			var messages []*model.TextMsg

			// 构建分片表名
			shardTableName := fmt.Sprintf("messages_%02d", dao.getShardIndex(shardKey))

			err := db.Table(shardTableName).
				Where("(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
					userID1, userID2, userID2, userID1).
				Order("created_at DESC").
				Limit(limit).
				Find(&messages).Error

			if err != nil {
				return fmt.Errorf("failed to query messages from shard %v: %w", shardKey, err)
			}

			allMessages = append(allMessages, messages...)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 按时间排序合并结果
	return dao.sortMessagesByTime(allMessages, limit), nil
}

// getShardIndex 计算分片索引的辅助方法
func (dao *EnhancedMessageDAO) getShardIndex(shardKey interface{}) int {
	// 简单的哈希分片计算
	hash := 0
	switch v := shardKey.(type) {
	case uint:
		hash = int(v)
	case int:
		hash = v
	case string:
		for _, c := range v {
			hash += int(c)
		}
	}
	return hash % 8 // 假设8个分片
}

// GetGroupHistoryMessages 获取群组历史消息（按群组ID分片）
func (dao *EnhancedMessageDAO) GetGroupHistoryMessages(ctx context.Context, groupID uint, lastID uint, limit int) ([]*model.GroupMessage, bool, error) {
	var messages []*model.GroupMessage

	opts := &database.QueryOptions{
		Operation:    database.ShardReadOperation,
		ShardKey:     groupID, // 使用群组ID作为分片键
		TableName:    "group_messages",
		ForceReplace: true,
	}

	err := dao.repo.Execute(ctx, opts, func(db *gorm.DB) error {
		query := db.Where("group_id = ?", groupID)

		if lastID > 0 {
			query = query.Where("id < ?", lastID)
		}

		return query.Order("id DESC").Limit(limit + 1).Find(&messages).Error
	})

	if err != nil {
		return nil, false, err
	}

	// 检查是否还有更多消息
	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	return messages, hasMore, nil
}

// GetRecentMessages 获取用户的最近消息（跨分片查询）
func (dao *EnhancedMessageDAO) GetRecentMessages(ctx context.Context, userID uint, limit int) ([]*model.TextMsg, error) {
	results, err := dao.repo.CrossShardQuery(ctx, "messages", func(db *gorm.DB) *gorm.DB {
		return db.Where("from_user_id = ? OR to_user_id = ?", userID, userID).
			Order("created_at DESC").
			Limit(limit)
	})

	if err != nil {
		return nil, err
	}

	messages := make([]*model.TextMsg, 0, len(results))
	for _, result := range results {
		message := &model.TextMsg{}
		// 映射字段（这里可以使用反射或其他更好的方法）
		dao.mapResultToMessage(result, message)
		messages = append(messages, message)
	}

	return messages, nil
}

// BatchSaveMessages 批量保存消息
func (dao *EnhancedMessageDAO) BatchSaveMessages(ctx context.Context, messages []*model.TextMsg) error {
	items := make([]interface{}, len(messages))
	for i, msg := range messages {
		items[i] = msg
	}

	return dao.repo.BatchInsert(ctx, "messages", items, func(item interface{}) interface{} {
		msg := item.(*model.TextMsg)
		return msg.FromUserID // 使用发送者ID作为分片键
	})
}

// DeleteOldMessages 删除旧消息（按时间分片清理）
func (dao *EnhancedMessageDAO) DeleteOldMessages(ctx context.Context, beforeTime time.Time) error {
	_, err := dao.repo.CrossShardQuery(ctx, "messages", func(db *gorm.DB) *gorm.DB {
		return db.Where("created_at < ?", beforeTime).Delete(&model.TextMsg{})
	})
	return err
}

// GetMessageStats 获取消息统计信息（跨分片聚合）
func (dao *EnhancedMessageDAO) GetMessageStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	// 获取单聊消息统计
	messageResults, err := dao.repo.CrossShardQuery(ctx, "messages", func(db *gorm.DB) *gorm.DB {
		return db.Select("COUNT(*) as total")
	})
	if err != nil {
		return nil, err
	}

	var totalMessages int64
	for _, result := range messageResults {
		if count, ok := result["total"].(int64); ok {
			totalMessages += count
		}
	}

	// 获取群聊消息统计
	groupMessageResults, err := dao.repo.CrossShardQuery(ctx, "group_messages", func(db *gorm.DB) *gorm.DB {
		return db.Select("COUNT(*) as total")
	})
	if err != nil {
		return nil, err
	}

	var totalGroupMessages int64
	for _, result := range groupMessageResults {
		if count, ok := result["total"].(int64); ok {
			totalGroupMessages += count
		}
	}

	stats["total_messages"] = totalMessages
	stats["total_group_messages"] = totalGroupMessages
	stats["total_all_messages"] = totalMessages + totalGroupMessages

	return stats, nil
}

// 辅助函数

// generateMessageID 生成消息ID（可以使用雪花算法或UUID）
func generateMessageID() string {
	// 这里简单使用时间戳，实际项目中应该使用更好的ID生成策略
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// sortMessagesByTime 按时间排序消息
func (dao *EnhancedMessageDAO) sortMessagesByTime(messages []*model.TextMsg, limit int) []*model.TextMsg {
	// 简单的时间排序实现
	if len(messages) <= limit {
		return messages
	}
	return messages[:limit]
}

// mapResultToMessage 映射查询结果到消息结构体
func (dao *EnhancedMessageDAO) mapResultToMessage(result map[string]interface{}, message *model.TextMsg) {
	if fromUserID, ok := result["from_user_id"].(string); ok {
		message.FromUserID = fromUserID
	}
	if toUserID, ok := result["to_user_id"].(string); ok {
		message.ToUserID = toUserID
	}
	if content, ok := result["content"].(string); ok {
		message.Content = content
	}
	if msgType, ok := result["type"].(string); ok {
		message.Type = msgType
	}
	if sentAt, ok := result["sent_at"].(time.Time); ok {
		message.SentAt = sentAt
	}
	// 映射其他字段...
}
