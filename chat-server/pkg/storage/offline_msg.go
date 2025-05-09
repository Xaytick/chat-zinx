package storage

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
)

// 全局变量存储离线消息
// TODO: 将离线消息存储到数据库中
var offlineMsgs = make(map[string][][]byte) // userID -> 消息列表(二进制格式)
var offlineMsgsMutex sync.RWMutex           // 并发保护

// SaveOfflineMsg 存储离线消息
func SaveOfflineMsg(userID string, msgData []byte) {
	offlineMsgsMutex.Lock()
	defer offlineMsgsMutex.Unlock()

	// 解析消息以便打印日志
	var msg model.TextMsg
	if err := json.Unmarshal(msgData, &msg); err == nil {
		fmt.Printf("[离线消息] 存储给用户ID=%s: %s\n", userID, msg.Content)
	} else {
		fmt.Printf("[离线消息] 存储给用户ID=%s: 消息解析失败: %v\n", userID, err)
	}

	// 存储原始二进制消息
	offlineMsgs[userID] = append(offlineMsgs[userID], msgData)

	printOfflineMessageStats()
}

// GetOfflineMessages 获取并清空用户的离线消息
func GetOfflineMessages(userID string) [][]byte {
	offlineMsgsMutex.Lock()
	defer offlineMsgsMutex.Unlock()

	msgs := offlineMsgs[userID]
	delete(offlineMsgs, userID) // 从离线消息列表中清空该用户的离线消息

	fmt.Printf("[离线消息] 用户ID=%s 获取离线消息，共 %d 条\n", userID, len(msgs))
	return msgs
}

// HasOfflineMessages 检查用户是否有离线消息
func HasOfflineMessages(userID string) bool {
	offlineMsgsMutex.RLock()
	defer offlineMsgsMutex.RUnlock()

	msgs, exists := offlineMsgs[userID]
	return exists && len(msgs) > 0
}

// Count 返回所有未读离线消息的数量
func Count() int {
	offlineMsgsMutex.RLock()
	defer offlineMsgsMutex.RUnlock()

	total := 0
	for _, msgs := range offlineMsgs {
		total += len(msgs)
	}
	return total
}

// printOfflineMessageStats 打印离线消息统计信息
func printOfflineMessageStats() {
	userCount := 0
	msgCount := 0

	for userID, msgs := range offlineMsgs {
		if len(msgs) > 0 {
			userCount++
			msgCount += len(msgs)
			fmt.Printf("[离线统计] 用户ID=%s: %d条消息\n", userID, len(msgs))
		}
	}

	fmt.Printf("[离线统计] 总计: %d个用户有%d条未读消息\n", userCount, msgCount)
}
