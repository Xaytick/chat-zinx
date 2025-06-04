// 分布式消息测试工具
// 用于验证跨服务器P2P和群组消息传递功能

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
)

type TestClient struct {
	conn       net.Conn
	serverID   string
	userID     uint
	username   string
	userUUID   string
	isLoggedIn bool
}

func NewTestClient(serverAddr, serverID string) *TestClient {
	return &TestClient{
		serverID: serverID,
	}
}

func (tc *TestClient) Connect() error {
	conn, err := net.Dial("tcp", tc.serverID)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}

	tc.conn = conn
	fmt.Printf("✓ 已连接到服务器 %s\n", tc.serverID)

	// 启动消息接收协程
	go tc.receiveMessages()

	return nil
}

func (tc *TestClient) receiveMessages() {
	for {
		if tc.conn == nil {
			break
		}

		// 读取消息头（8字节：4字节长度 + 4字节消息ID）
		headerData := make([]byte, 8)
		_, err := tc.conn.Read(headerData)
		if err != nil {
			fmt.Printf("读取消息头失败: %v\n", err)
			break
		}

		// 解析消息头
		dataLen := uint32(headerData[0]) | uint32(headerData[1])<<8 |
			uint32(headerData[2])<<16 | uint32(headerData[3])<<24
		msgID := uint32(headerData[4]) | uint32(headerData[5])<<8 |
			uint32(headerData[6])<<16 | uint32(headerData[7])<<24

		// 读取消息体
		if dataLen > 0 {
			msgData := make([]byte, dataLen)
			_, err := tc.conn.Read(msgData)
			if err != nil {
				fmt.Printf("读取消息体失败: %v\n", err)
				break
			}

			tc.handleMessage(msgID, msgData)
		}
	}
}

func (tc *TestClient) handleMessage(msgID uint32, data []byte) {
	switch msgID {
	case protocol.MsgIDLoginResp:
		tc.handleLoginResponse(data)
	case protocol.MsgIDTextMsg:
		tc.handleTextMessage(data)
	case protocol.MsgIDGroupTextMsgResp:
		tc.handleGroupMessage(data)
	default:
		fmt.Printf("收到未知消息类型: %d\n", msgID)
	}
}

func (tc *TestClient) handleLoginResponse(data []byte) {
	var resp model.LoginResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		fmt.Printf("解析登录响应失败: %v\n", err)
		return
	}

	if resp.Code == 200 {
		tc.isLoggedIn = true
		tc.userID = resp.UserID
		tc.username = resp.Username
		tc.userUUID = resp.UserUUID
		fmt.Printf("✓ 登录成功! 用户: %s (ID: %d, UUID: %s)\n",
			tc.username, tc.userID, tc.userUUID)
	} else {
		fmt.Printf("✗ 登录失败: %s\n", resp.Message)
	}
}

func (tc *TestClient) handleTextMessage(data []byte) {
	var msg model.TextMsg
	if err := json.Unmarshal(data, &msg); err != nil {
		fmt.Printf("解析文本消息失败: %v\n", err)
		return
	}

	fmt.Printf("📨 [P2P消息] 来自 %s: %s\n", msg.FromUsername, msg.Content)
}

func (tc *TestClient) handleGroupMessage(data []byte) {
	var msg model.GroupTextMsg
	if err := json.Unmarshal(data, &msg); err != nil {
		fmt.Printf("解析群组消息失败: %v\n", err)
		return
	}

	fmt.Printf("📢 [群组消息] 群组%d - %s: %s\n",
		msg.GroupID, msg.FromUsername, msg.Content)
}

func (tc *TestClient) sendMessage(msgID uint32, data []byte) error {
	// 构造消息包
	dataLen := uint32(len(data))

	// 构造消息头（8字节）
	header := make([]byte, 8)
	header[0] = byte(dataLen)
	header[1] = byte(dataLen >> 8)
	header[2] = byte(dataLen >> 16)
	header[3] = byte(dataLen >> 24)
	header[4] = byte(msgID)
	header[5] = byte(msgID >> 8)
	header[6] = byte(msgID >> 16)
	header[7] = byte(msgID >> 24)

	// 发送消息头
	_, err := tc.conn.Write(header)
	if err != nil {
		return err
	}

	// 发送消息体
	if len(data) > 0 {
		_, err = tc.conn.Write(data)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tc *TestClient) Login(username, password string) error {
	loginReq := model.LoginRequest{
		Username: username,
		Password: password,
	}

	data, err := json.Marshal(loginReq)
	if err != nil {
		return err
	}

	return tc.sendMessage(protocol.MsgIDLogin, data)
}

func (tc *TestClient) SendTextMessage(toUserID string, content string) error {
	if !tc.isLoggedIn {
		return fmt.Errorf("请先登录")
	}

	// 尝试将toUserID转换为uint
	toUserIDUint, err := strconv.ParseUint(toUserID, 10, 32)
	if err != nil {
		// 如果不是数字，可能是UUID，直接使用
		textMsg := model.TextMsg{
			FromUserID:   tc.userID,
			FromUsername: tc.username,
			ToUserUUID:   toUserID,
			Content:      content,
			SendTime:     time.Now().Unix(),
		}

		data, err := json.Marshal(textMsg)
		if err != nil {
			return err
		}

		return tc.sendMessage(protocol.MsgIDTextMsg, data)
	}

	textMsg := model.TextMsg{
		FromUserID:   tc.userID,
		FromUsername: tc.username,
		ToUserID:     uint(toUserIDUint),
		Content:      content,
		SendTime:     time.Now().Unix(),
	}

	data, err := json.Marshal(textMsg)
	if err != nil {
		return err
	}

	return tc.sendMessage(protocol.MsgIDTextMsg, data)
}

func (tc *TestClient) SendGroupMessage(groupID uint, content string) error {
	if !tc.isLoggedIn {
		return fmt.Errorf("请先登录")
	}

	groupMsg := model.GroupTextMsg{
		GroupID:      groupID,
		FromUserID:   tc.userID,
		FromUsername: tc.username,
		Content:      content,
		SendTime:     time.Now().Unix(),
	}

	data, err := json.Marshal(groupMsg)
	if err != nil {
		return err
	}

	return tc.sendMessage(protocol.MsgIDGroupTextMsg, data)
}

func (tc *TestClient) Close() {
	if tc.conn != nil {
		tc.conn.Close()
		tc.conn = nil
	}
}

func main() {
	fmt.Println("=== Chat-Zinx 分布式消息测试工具 ===")
	fmt.Println("此工具用于测试跨服务器P2P和群组消息传递")
	fmt.Println()

	// 服务器地址列表
	servers := []string{
		"localhost:9000",
		"localhost:9001",
		"localhost:9002",
	}

	// 显示可用服务器
	fmt.Println("可用服务器:")
	for i, server := range servers {
		fmt.Printf("  %d. %s\n", i+1, server)
	}
	fmt.Println()

	// 选择服务器
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("请选择连接的服务器 (1-3): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	serverIndex, err := strconv.Atoi(input)
	if err != nil || serverIndex < 1 || serverIndex > len(servers) {
		log.Fatal("无效的服务器选择")
	}

	selectedServer := servers[serverIndex-1]

	// 创建客户端
	client := NewTestClient(selectedServer, selectedServer)

	// 连接服务器
	if err := client.Connect(); err != nil {
		log.Fatalf("连接服务器失败: %v", err)
	}
	defer client.Close()

	// 登录
	fmt.Print("用户名: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("密码: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	if err := client.Login(username, password); err != nil {
		log.Fatalf("发送登录请求失败: %v", err)
	}

	// 等待登录响应
	time.Sleep(2 * time.Second)

	if !client.isLoggedIn {
		log.Fatal("登录失败")
	}

	// 交互式命令循环
	fmt.Println("\n=== 命令帮助 ===")
	fmt.Println("  p <用户ID/UUID> <消息>  - 发送P2P消息")
	fmt.Println("  g <群组ID> <消息>       - 发送群组消息")
	fmt.Println("  quit                   - 退出程序")
	fmt.Println()

	for {
		fmt.Printf("[%s@%s] > ", client.username, selectedServer)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "quit" || input == "exit" {
			break
		}

		parts := strings.SplitN(input, " ", 3)
		if len(parts) < 3 {
			fmt.Println("命令格式错误，请查看帮助")
			continue
		}

		command := parts[0]
		target := parts[1]
		message := parts[2]

		switch command {
		case "p":
			if err := client.SendTextMessage(target, message); err != nil {
				fmt.Printf("发送P2P消息失败: %v\n", err)
			} else {
				fmt.Printf("✓ P2P消息已发送到 %s\n", target)
			}

		case "g":
			groupID, err := strconv.ParseUint(target, 10, 32)
			if err != nil {
				fmt.Printf("无效的群组ID: %s\n", target)
				continue
			}

			if err := client.SendGroupMessage(uint(groupID), message); err != nil {
				fmt.Printf("发送群组消息失败: %v\n", err)
			} else {
				fmt.Printf("✓ 群组消息已发送到群组 %d\n", groupID)
			}

		default:
			fmt.Println("未知命令，请查看帮助")
		}
	}

	fmt.Println("感谢使用分布式消息测试工具!")
}
