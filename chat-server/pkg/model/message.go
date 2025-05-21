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

// LegacyHistoryMsgReq 旧版获取历史消息请求结构
// 由客户端发送给服务端
type LegacyHistoryMsgReq struct {
	TargetUserUUID string `json:"target_user_uuid,omitempty"` // 目标用户UUID (可选, 与谁的聊天历史)
	TargetUsername string `json:"target_username,omitempty"`  // 目标用户名 (可选, 与谁的聊天历史)
	Limit          int    `json:"limit"`                      // 获取消息的数量限制
}

// LegacyHistoryMsgResp 旧版历史消息响应结构
// 服务端返回给客户端
type LegacyHistoryMsgResp struct {
	Code    uint32                   `json:"code"`
	Message string                   `json:"message"`
	Data    []map[string]interface{} `json:"data,omitempty"` // 历史消息列表, 每一项是 map[string]interface{}
}

// GenericMessageResp 通用消息响应，常用于操作成功/失败的简单反馈
type GenericMessageResp struct {
	Code    uint32 `json:"code"`
	Message string `json:"message"`
}

// TextMsgReq 文本消息请求
type TextMsgReq struct {
	ToUserID uint   `json:"to_user_id"` // 接收者ID
	Content  string `json:"content"`    // 文本消息内容
}

// TextMsgResp 文本消息响应
type TextMsgResp struct {
	MsgID  string `json:"msg_id,omitempty"` // 消息的唯一ID (可选, 便于客户端追踪)
	Status int32  `json:"status"`           // 状态码, 0表示成功, 其他表示错误
	Error  string `json:"error,omitempty"`  // 错误信息, 成功时为空
}

// TextMsgPush 推送的文本消息
type TextMsgPush struct {
	FromUserID   uint   `json:"from_user_id"`   // 发送者用户ID
	FromUserUUID string `json:"from_user_uuid"` // 发送者UUID
	FromUsername string `json:"from_username"`  // 发送者用户名
	Content      string `json:"content"`        // 消息内容
	Timestamp    int64  `json:"timestamp"`      // 服务器收到消息时的时间戳 (Unix秒)
}

// HistoryMsgReq 历史消息请求
type HistoryMsgReq struct {
	PeerUserID uint   `json:"peer_user_id"` // 对方用户ID
	LastMsgID  string `json:"last_msg_id"`  // 上一条消息的ID，用于分页查询
	Limit      int    `json:"limit"`        // 需要获取的消息数量
}

// MessageItemResp 消息项结构（用于历史消息响应）
type MessageItemResp struct {
	MsgID        string `json:"msg_id"`         // 消息ID
	FromUserID   uint   `json:"from_user_id"`   // 发送者ID
	FromUserUUID string `json:"from_user_uuid"` // 发送者UUID
	FromUsername string `json:"from_username"`  // 发送者名称
	Content      string `json:"content"`        // 消息内容
	MsgType      string `json:"msg_type"`       // 消息类型，如 "text", "image", "file" 等
	Timestamp    int64  `json:"timestamp"`      // 消息创建时间戳
}

// HistoryMsgResp 历史消息响应
type HistoryMsgResp struct {
	Messages []*MessageItemResp `json:"messages"` // 消息列表
	HasMore  bool               `json:"has_more"` // 是否还有更多消息
}
