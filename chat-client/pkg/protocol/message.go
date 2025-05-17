package protocol

// Message 客户端和服务端通信的消息结构
type Message struct {
	DataLen uint32 // 消息长度
	ID      uint32 // 消息ID
	Data    []byte // 消息内容
}

// GetDataLen 获取消息数据段长度
func (m *Message) GetDataLen() uint32 {
	return m.DataLen
}

// GetMsgID 获取消息ID
func (m *Message) GetMsgID() uint32 {
	return m.ID
}

// GetData 获取消息内容
func (m *Message) GetData() []byte {
	return m.Data
}

// SetDataLen 设置消息数据段长度
func (m *Message) SetDataLen(len uint32) {
	m.DataLen = len
}

// SetMsgID 设置消息ID
func (m *Message) SetMsgID(msgID uint32) {
	m.ID = msgID
}

// SetData 设置消息内容
func (m *Message) SetData(data []byte) {
	m.Data = data
}
