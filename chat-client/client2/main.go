package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Xaytick/chat-zinx/chat-client/pkg/client"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
)

func main() {
	// 创建客户端实例
	cli, err := client.NewChatClient("127.0.0.1:9000")
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	// 注册并登录用户
	if err := cli.RegisterAndLogin("testuser2", "123456"); err != nil {
		panic(err)
	}

	// 等待一会，确保连接完全建立
	time.Sleep(500 * time.Millisecond)

	// 发送消息给 testuser1
	targetUsername := "testuser1"
	fmt.Printf("发送消息给用户: %s\n", targetUsername)
	if err := cli.SendTextMessage(targetUsername, "你好, testuser1! 我是testuser2!"); err != nil {
		fmt.Printf("发送消息失败: %v\n", err)
	} else {
		fmt.Println("消息发送成功！")
	}

	// 处理接收到的消息
	fmt.Println("等待接收消息...")
	cli.StartMsgListener(func(msgID uint32, msgBody []byte) {
		switch msgID {
		case protocol.MsgIDTextMsg:
			var msg map[string]interface{}
			json.Unmarshal(msgBody, &msg)
			fmt.Printf("\n收到消息: %v\n", msg)
			if content, ok := msg["content"].(string); ok {
				fmt.Printf("消息内容: %s\n", content)
			}
		default:
			fmt.Printf("\n收到消息 ID=%d, 内容=%s\n", msgID, string(msgBody))
		}
	})

	// 阻塞主程序
	select {}
}
