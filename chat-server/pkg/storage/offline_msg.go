package storage

import (
	"fmt"
	"sync"
)

// 全局变量存储离线消息
// TODO: 将离线消息存储到数据库中
var offlineMsgs = make(map[string][]string) // userID -> 消息列表
var offlineMsgsMutex sync.RWMutex           // 并发保护

// SaveOfflineMsg 存储离线消息
func SaveOfflineMsg(userID string, msgContent string) {
	offlineMsgsMutex.Lock()
	defer offlineMsgsMutex.Unlock()

	offlineMsgs[userID] = append(offlineMsgs[userID], msgContent)
	fmt.Printf("存储离线消息: 给用户 %s, 内容: %s\n", userID, msgContent)
}

// GetOfflineMessages 获取并清空用户的离线消息
func GetOfflineMessages(userID string) []string {
	offlineMsgsMutex.Lock()
	defer offlineMsgsMutex.Unlock()
	
	msgs := offlineMsgs[userID]
	delete(offlineMsgs, userID) // 从离线消息列表中清空该用户的离线消息
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