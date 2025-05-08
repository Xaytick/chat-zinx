package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:9000")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// 构造登录请求
	req := map[string]string{
		"username": "testuser1",
		"password": "123456",
	}
	body, _ := json.Marshal(req)
	msgID := protocol.MsgIDLoginReq // MsgIDLoginReq
	length := uint32(len(body))

	// 组包
	buf := make([]byte, 8+len(body))
	binary.LittleEndian.PutUint32(buf[0:4], length)
	binary.LittleEndian.PutUint32(buf[4:8], msgID)
	copy(buf[8:], body)

	// 发送
	conn.Write(buf)

	// 读取登录响应
	head := make([]byte, 8)
	conn.Read(head)
	respLen := binary.LittleEndian.Uint32(head[0:4])
	respBody := make([]byte, respLen)
	conn.Read(respBody)
	fmt.Println("服务器响应：", string(respBody))

	// 构造单聊消息
	msg := map[string]interface{}{
		"to_user_id": "testuser2", // 目标用户ID（需和服务端绑定一致）
		"content":    "你好！",
	}
	msgBody, _ := json.Marshal(msg)
	msgID = protocol.MsgIDTextMsg // 从 chat-server 的 protocol 包引入
	length = uint32(len(msgBody))

	// 组包
	msgBuf := make([]byte, 8+len(msgBody))
	binary.LittleEndian.PutUint32(msgBuf[0:4], length)
	binary.LittleEndian.PutUint32(msgBuf[4:8], msgID)
	copy(msgBuf[8:], msgBody)

	// 发送单聊消息
	conn.Write(msgBuf)

	// 读取单聊响应（如果有）
	conn.Read(head)
	respLen = binary.LittleEndian.Uint32(head[0:4])
	respBody = make([]byte, respLen)
	conn.Read(respBody)
	fmt.Println("服务器响应：", string(respBody))
}
