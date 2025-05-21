package model

import "time"

// GroupTextMsgReq C->S 发送群组文本消息请求
type GroupTextMsgReq struct {
	GroupID uint32 `json:"group_id"` // 群组ID
	Content string `json:"content"`  // 消息内容
}

// GroupTextMsgResp S->C 发送群组文本消息响应
type GroupTextMsgResp struct {
	MsgID  string `json:"msg_id,omitempty"` // 消息的唯一ID (可选, 便于客户端追踪)
	Status int32  `json:"status"`           // 状态码, 0表示成功, 其他表示错误
	Error  string `json:"error,omitempty"`  // 错误信息, 成功时为空
}

// GroupTextMsgPush S->C 推送群组文本消息
type GroupTextMsgPush struct {
	GroupID      uint32 `json:"group_id"`       // 群组ID
	FromUserID   uint   `json:"from_user_id"`   // 发送者DB User ID
	FromUserUUID string `json:"from_user_uuid"` // 发送者User UUID
	FromUsername string `json:"from_username"`  // 发送者用户名
	Content      string `json:"content"`        // 消息内容
	Timestamp    int64  `json:"timestamp"`      // 服务器收到消息时的时间戳 (Unix秒)
}

// GroupMessage 群组消息数据库存储模型
type GroupMessage struct {
	ID          uint      `json:"id" gorm:"primarykey"`               // 消息ID
	MsgID       string    `json:"msg_id" gorm:"type:varchar(36)"`     // 消息唯一标识
	GroupID     uint      `json:"group_id" gorm:"index:idx_group_id"` // 群组ID，用于查询
	SenderID    uint      `json:"sender_id"`                          // 发送者用户ID
	SenderUUID  string    `json:"sender_uuid"`                        // 发送者UUID
	SenderName  string    `json:"sender_name"`                        // 发送者用户名
	Content     string    `json:"content" gorm:"type:text"`           // 消息内容
	MessageType string    `json:"message_type" gorm:"default:'text'"` // 消息类型: text, image, file 等
	CreatedAt   time.Time `json:"created_at" gorm:"index"`            // 创建时间
}

// GroupHistoryMsgReq 获取群组历史消息请求
type GroupHistoryMsgReq struct {
	GroupID uint `json:"group_id"`          // 群组ID
	LastID  uint `json:"last_id,omitempty"` // 上次查询的最后一条消息ID，用于分页
	Limit   int  `json:"limit,omitempty"`   // 查询数量限制
}

// GroupHistoryMsgItem 群组历史消息单条数据
type GroupHistoryMsgItem struct {
	ID          uint   `json:"id"`           // 消息ID
	MsgID       string `json:"msg_id"`       // 消息唯一标识
	SenderID    uint   `json:"sender_id"`    // 发送者ID
	SenderUUID  string `json:"sender_uuid"`  // 发送者UUID
	SenderName  string `json:"sender_name"`  // 发送者名称
	Content     string `json:"content"`      // 消息内容
	MessageType string `json:"message_type"` // 消息类型
	Timestamp   int64  `json:"timestamp"`    // 时间戳（Unix秒）
}

// GroupHistoryMsgResp 获取群组历史消息响应
type GroupHistoryMsgResp struct {
	GroupID  uint                   `json:"group_id"` // 群组ID
	Messages []*GroupHistoryMsgItem `json:"messages"` // 消息列表
	HasMore  bool                   `json:"has_more"` // 是否还有更多消息
}
