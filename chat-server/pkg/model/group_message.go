package model

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
