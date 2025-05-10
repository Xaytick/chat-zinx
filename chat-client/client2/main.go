package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Xaytick/chat-zinx/chat-client/pkg/client"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
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
			// 尝试直接解析JSON
			var msg model.TextMsg
			if err := json.Unmarshal(msgBody, &msg); err == nil {
				fromUsername := cli.GetUsernameByID(msg.FromUserID)
				fmt.Printf("\n收到来自 %s 的消息: %s\n", fromUsername, msg.Content)
			} else {
				// 检查是否是Base64编码
				msgStr := string(msgBody)
				if isBase64(msgStr) {
					decodedBytes, err := base64.StdEncoding.DecodeString(msgStr)
					if err == nil {
						// 尝试解析解码后的JSON
						if err := json.Unmarshal(decodedBytes, &msg); err == nil {
							fromUsername := cli.GetUsernameByID(msg.FromUserID)
							fmt.Printf("\n收到来自 %s 的Base64编码消息: %s\n", fromUsername, msg.Content)
						} else {
							fmt.Printf("\n解析Base64解码后的消息失败: %v, 内容: %s\n", err, string(decodedBytes))
						}
					} else {
						fmt.Printf("\nBase64解码失败: %v, 原始内容: %s\n", err, msgStr)
					}
				} else {
					fmt.Printf("\n解析消息失败: %v, 原始内容: %s\n", err, string(msgBody))
				}
			}
		default:
			fmt.Printf("\n收到消息 ID=%d, 内容=%s\n", msgID, string(msgBody))
		}
	})

	// 阻塞主程序
	select {}
}

// isBase64 检查字符串是否可能是Base64编码
func isBase64(s string) bool {
	// 简单检查: Base64字符串通常由[A-Za-z0-9+/=]组成，长度是4的倍数
	// 并且通常以'e'开头(对应JSON的{)
	return len(s) > 0 && len(s)%4 == 0 && s[0] == 'e' && s[1] == 'y'
}
