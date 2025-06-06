package storage

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/Xaytick/chat-zinx/chat-server/conf"
	"github.com/Xaytick/chat-zinx/chat-server/dao/redis"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
)

const (
	// 消息列表的键前缀
	offlineMessagePrefix = "offline:msg:"
	// 历史消息的键前缀
	historyMessagePrefix = "history:msg:"
	// 双向消息关系的键前缀
	chatRelationPrefix = "chat:relation:"
)

// RedisMsgStorage Redis实现的消息存储
type RedisMsgStorage struct {
	expiration time.Duration // 消息过期时间
}

// NewRedisMsgStorage 创建Redis消息存储
func NewRedisMsgStorage() *RedisMsgStorage {
	cfg := conf.GlobalConfig.Database.Redis
	return &RedisMsgStorage{
		expiration: redis.GetMessageExpiration(&cfg),
	}
}

// generateOfflineKey 生成离线消息的键
func generateOfflineKey(userID uint) string {
	return offlineMessagePrefix + strconv.FormatUint(uint64(userID), 10)
}

// generateHistoryKey 生成历史消息的键
func generateHistoryKey(fromUserID, toUserID uint) string {
	fromStr := strconv.FormatUint(uint64(fromUserID), 10)
	toStr := strconv.FormatUint(uint64(toUserID), 10)
	// 确保两个用户ID的顺序是固定的，不管是谁发给谁
	if fromUserID < toUserID {
		return historyMessagePrefix + fromStr + ":" + toStr
	}
	return historyMessagePrefix + toStr + ":" + fromStr
}

// generateRelationKey 生成聊天关系的键
func generateRelationKey(userID uint) string {
	return chatRelationPrefix + strconv.FormatUint(uint64(userID), 10)
}

// SaveMessage 存储消息
// 同时保存到离线消息队列和历史消息列表
func (s *RedisMsgStorage) SaveMessage(fromUserID, toUserID uint, msgData []byte) error {
	fromUserIDStr := strconv.FormatUint(uint64(fromUserID), 10)
	toUserIDStr := strconv.FormatUint(uint64(toUserID), 10)

	// 解析消息以便打印日志
	var msg model.TextMsg
	if err := json.Unmarshal(msgData, &msg); err == nil {
		fmt.Printf("[Redis消息] 存储消息: 从 %s -> %s: %s\n", fromUserIDStr, toUserIDStr, msg.Content)
	}

	// 构建消息元数据
	msgMeta := map[string]interface{}{
		"from_user_id": fromUserIDStr,
		"to_user_id":   toUserIDStr,
		"data":         msgData,
		"timestamp":    time.Now().Unix(),
	}

	// 序列化消息元数据
	msgMetaJson, err := json.Marshal(msgMeta)
	if err != nil {
		return err
	}

	// 保存到离线消息队列 (用于接收方上线后推送)
	offlineKey := generateOfflineKey(toUserID)
	err = redis.RedisClient.RPush(redis.Ctx, offlineKey, msgMetaJson).Err()
	if err != nil {
		return err
	}
	// 设置过期时间
	redis.RedisClient.Expire(redis.Ctx, offlineKey, s.expiration)

	// 保存到历史消息列表 (双方查看聊天记录)
	historyKey := generateHistoryKey(fromUserID, toUserID)
	err = redis.RedisClient.RPush(redis.Ctx, historyKey, msgMetaJson).Err()
	if err != nil {
		return err
	}
	// 设置过期时间
	redis.RedisClient.Expire(redis.Ctx, historyKey, s.expiration)

	// 更新聊天关系列表 (记录与谁有过聊天)
	// 发送方的关系
	relationKey1 := generateRelationKey(fromUserID)
	redis.RedisClient.SAdd(redis.Ctx, relationKey1, toUserIDStr)
	redis.RedisClient.Expire(redis.Ctx, relationKey1, s.expiration)

	// 接收方的关系
	relationKey2 := generateRelationKey(toUserID)
	redis.RedisClient.SAdd(redis.Ctx, relationKey2, fromUserIDStr)
	redis.RedisClient.Expire(redis.Ctx, relationKey2, s.expiration)

	return nil
}

// GetOfflineMessages 获取并清空用户的离线消息
func (s *RedisMsgStorage) GetOfflineMessages(userID uint) ([][]byte, error) {
	offlineKey := generateOfflineKey(userID)

	// 使用Lua脚本实现原子性操作：获取所有消息并删除键
	script := `
	local messages = redis.call('LRANGE', KEYS[1], 0, -1)
	redis.call('DEL', KEYS[1])
	return messages
	`
	result, err := redis.RedisClient.Eval(redis.Ctx, script, []string{offlineKey}).Result()
	if err != nil {
		return nil, err
	}

	results, ok := result.([]interface{})
	if !ok || len(results) == 0 {
		return [][]byte{}, nil
	}

	// 转换为二进制数据
	messages := make([][]byte, 0, len(results))
	for _, result := range results {
		resultStr, ok := result.(string)
		if !ok {
			continue
		}

		var msgMeta map[string]interface{}
		if err := json.Unmarshal([]byte(resultStr), &msgMeta); err != nil {
			fmt.Printf("[Redis消息] 解析离线消息元数据失败: %v\n", err)
			continue
		}

		// 提取消息数据
		if dataBytes, ok := msgMeta["data"].([]byte); ok {
			messages = append(messages, dataBytes)
		} else if dataStr, ok := msgMeta["data"].(string); ok {
			// 注意：Redis存储的字符串可能是Base64或JSON字符串形式
			// 尝试直接解析JSON
			var msgObj model.TextMsg
			if err := json.Unmarshal([]byte(dataStr), &msgObj); err == nil {
				// 确保消息包含发送者ID
				if msgObj.FromUserID == "" {
					// 如果消息中没有发送者ID，从元数据中添加
					msgObj.FromUserID = msgMeta["from_user_id"].(string)
					// 重新序列化包含发送者ID的消息
					enhancedData, err := json.Marshal(msgObj)
					if err == nil {
						messages = append(messages, enhancedData)
					} else {
						messages = append(messages, []byte(dataStr))
					}
				} else {
					// 消息已包含发送者ID，直接使用
					messages = append(messages, []byte(dataStr))
				}
			} else {
				// 如果不是有效的JSON，直接使用原始数据
				messages = append(messages, []byte(dataStr))
			}
		}
	}

	if len(messages) > 0 {
		fmt.Printf("[Redis消息] 用户ID=%s 获取离线消息，共 %d 条\n", strconv.FormatUint(uint64(userID), 10), len(messages))
	}

	return messages, nil
}

// HasOfflineMessages 检查用户是否有离线消息
func (s *RedisMsgStorage) HasOfflineMessages(userID uint) (bool, error) {
	offlineKey := generateOfflineKey(userID)
	count, err := redis.RedisClient.LLen(redis.Ctx, offlineKey).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetHistoryMessages 获取历史消息
func (s *RedisMsgStorage) GetHistoryMessages(userID1, userID2 uint, limit int64) ([]map[string]interface{}, error) {
	historyKey := generateHistoryKey(userID1, userID2)

	results, err := redis.RedisClient.LRange(redis.Ctx, historyKey, -limit, -1).Result()
	if err != nil {
		fmt.Printf("[Redis历史-查询] Error on LRange for key %s: %v\n", historyKey, err)
		return nil, err
	}

	if len(results) == 0 {
		fmt.Printf("[Redis历史-查询] No messages found in Redis for key %s.\n", historyKey)
	}

	messagesForClient := make([]map[string]interface{}, 0, len(results))
	for _, resultStr := range results { // Removed index 'i' as it's no longer used for logging
		var msgMeta map[string]interface{}
		if err := json.Unmarshal([]byte(resultStr), &msgMeta); err != nil {
			// Keep this log as it indicates a data corruption/parsing issue
			fmt.Printf("[RedisStorage] GetHistoryMessages: Failed to unmarshal msgMeta: %v. Original: %s\n", err, resultStr)
			continue
		}

		originalMsgData, ok := msgMeta["data"].(string)
		if !ok {
			if dataBytes, okBytes := msgMeta["data"].([]interface{}); okBytes {
				tempBytes := make([]byte, len(dataBytes))
				for idx, v := range dataBytes {
					tempBytes[idx] = byte(v.(float64))
				}
				originalMsgData = string(tempBytes)
			} else {
				// Keep this log as it indicates a data structure issue
				fmt.Printf("[RedisStorage] GetHistoryMessages: 'data' field type is unexpected or missing in msgMeta: %+v. Skipping.\n", msgMeta)
				continue
			}
		}

		decodedData, err := base64.StdEncoding.DecodeString(originalMsgData)
		if err != nil {
			// Keep this log as it indicates a data corruption/decoding issue
			fmt.Printf("[RedisStorage] GetHistoryMessages: Failed to Base64 decode originalMsgData: %v. Original data string: %s. Skipping.\n", err, originalMsgData)
			continue
		}

		var textMsgPayload model.TextMsg
		if err := json.Unmarshal(decodedData, &textMsgPayload); err != nil {
			// Keep this log as it indicates a data corruption/parsing issue
			fmt.Printf("[RedisStorage] GetHistoryMessages: Failed to unmarshal TextMsg from decodedData: %v. Decoded data string: %s. Skipping.\n", err, string(decodedData))
			continue
		}

		clientMap := map[string]interface{}{
			"from_user_id": msgMeta["from_user_id"],
			"content":      textMsgPayload.Content,
			"timestamp":    msgMeta["timestamp"],
		}
		messagesForClient = append(messagesForClient, clientMap)
	}

	return messagesForClient, nil
}

// GetChatRelations 获取用户的聊天关系列表
func (s *RedisMsgStorage) GetChatRelations(userID uint) ([]string, error) {
	relationKey := generateRelationKey(userID)
	return redis.RedisClient.SMembers(redis.Ctx, relationKey).Result()
}

// SaveHistoryOnly 只保存历史消息，不保存离线消息
// 用于当消息已直接发送给在线用户时
func (s *RedisMsgStorage) SaveHistoryOnly(fromUserID, toUserID uint, msgData []byte) error {
	fromUserIDStr := strconv.FormatUint(uint64(fromUserID), 10)
	toUserIDStr := strconv.FormatUint(uint64(toUserID), 10)
	// 解析消息以便打印日志
	var msg model.TextMsg
	if err := json.Unmarshal(msgData, &msg); err == nil {
		fmt.Printf("[Redis历史] 只保存历史消息: 从 %s -> %s: %s\n", fromUserIDStr, toUserIDStr, msg.Content)
	}

	// 构建消息元数据
	msgMeta := map[string]interface{}{
		"from_user_id": fromUserIDStr,
		"to_user_id":   toUserIDStr,
		"data":         msgData,
		"timestamp":    time.Now().Unix(),
	}

	// 序列化消息元数据
	msgMetaJson, err := json.Marshal(msgMeta)
	if err != nil {
		return err
	}

	// 只保存到历史消息列表 (双方查看聊天记录)
	historyKey := generateHistoryKey(fromUserID, toUserID)
	err = redis.RedisClient.RPush(redis.Ctx, historyKey, msgMetaJson).Err()
	if err != nil {
		return err
	}
	// 设置过期时间
	redis.RedisClient.Expire(redis.Ctx, historyKey, s.expiration)

	// 更新聊天关系列表 (记录与谁有过聊天)
	// 发送方的关系
	relationKey1 := generateRelationKey(fromUserID)
	redis.RedisClient.SAdd(redis.Ctx, relationKey1, toUserIDStr)
	redis.RedisClient.Expire(redis.Ctx, relationKey1, s.expiration)

	// 接收方的关系
	relationKey2 := generateRelationKey(toUserID)
	redis.RedisClient.SAdd(redis.Ctx, relationKey2, fromUserIDStr)
	redis.RedisClient.Expire(redis.Ctx, relationKey2, s.expiration)

	return nil
}
