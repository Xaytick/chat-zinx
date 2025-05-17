package model

import "time"

// TextMsg 文本消息结构
// 服务端在收到消息后，会根据连接信息填充 FromUserID (通常为 UserUUID)。
// 客户端发送时，ToUserID 可以是目标用户的 UserUUID，或者是群组的 GroupID (string 格式)。
// 服务端根据 ToUserID 的格式或是否存在于群组列表来判断是私聊还是群聊。
// Type 字段 (e.g., "private", "group") 可由服务端补充，或客户端指定。
// SentAt 由服务端设置。
type TextMsg struct {
	FromUserID string    `json:"from_user_id,omitempty"` // 发送者ID (UserUUID), 服务端可覆盖/填充
	ToUserID   string    `json:"to_user_id"`             // 接收者ID (UserUUID 或 GroupID)
	Content    string    `json:"content"`                // 消息内容
	Type       string    `json:"type,omitempty"`         // 消息类型: private, group
	SentAt     time.Time `json:"sent_at,omitempty"`      // 发送时间 (服务端设置)
}

// HistoryMsgReq 获取历史消息请求结构
// 由客户端发送给服务端
type HistoryMsgReq struct {
	TargetUserUUID string `json:"target_user_uuid"` // 目标用户UUID (与谁的聊天历史)
	Limit          int    `json:"limit"`            // 获取消息的数量限制
}

// HistoryMsgResp 历史消息响应结构
// 服务端返回给客户端
type HistoryMsgResp struct {
	Code    uint32                   `json:"code"`
	Message string                   `json:"message"`
	Data    []map[string]interface{} `json:"data,omitempty"` // 历史消息列表, 每一项是 map[string]interface{}
}

// GenericMessageResp 通用消息响应，常用于操作成功/失败的简单反馈
type GenericMessageResp struct {
	Code    uint32 `json:"code"`
	Message string `json:"message"`
}
