package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	clientProtocol "github.com/Xaytick/chat-zinx/chat-client/pkg/protocol" // Client's own protocol for Message/DataPack
	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	serverProtocol "github.com/Xaytick/chat-zinx/chat-server/pkg/protocol" // Alias for server's protocol constants
)

// ChatClient 聊天客户端结构体
type ChatClient struct {
	Conn       net.Conn
	ServerAddr string
	UserID     uint   // User's primary key ID
	UserUUID   string // User's UUID
	Username   string // User's username
	Token      string // JWT Token

	isLoggedIn    bool
	heartbeatStop chan struct{}
	msgHandler    func(msgID uint32, data []byte) // Callback for received messages
}

// NewChatClient 创建一个新的聊天客户端
func NewChatClient(serverAddr string) (*ChatClient, error) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("连接服务器失败: %v", err)
	}

	return &ChatClient{
		Conn:          conn,
		ServerAddr:    serverAddr,
		heartbeatStop: make(chan struct{}),
	}, nil
}

// Close 关闭客户端连接
func (c *ChatClient) Close() {
	c.StopHeartbeat()
	if c.Conn != nil {
		c.Conn.Close()
	}
	c.isLoggedIn = false
}

// SendMessage 封装了消息的打包和发送过程
func (c *ChatClient) SendMessage(msgID uint32, data []byte) error {
	if c.Conn == nil {
		return errors.New("connection is not established")
	}
	msg := &clientProtocol.Message{ // Use clientProtocol.Message
		DataLen: uint32(len(data)),
		ID:      msgID,
		Data:    data,
	}

	dp := clientProtocol.NewDataPack() // Use clientProtocol.NewDataPack
	packedMsg, err := dp.Pack(msg)
	if err != nil {
		return fmt.Errorf("failed to pack message: %w", err)
	}

	_, err = c.Conn.Write(packedMsg)
	return err
}

// readMessage 读取并解包一个完整的消息
func (c *ChatClient) readMessage() (*clientProtocol.Message, error) { // Return clientProtocol.Message
	if c.Conn == nil {
		return nil, errors.New("connection is not established")
	}
	dp := clientProtocol.NewDataPack() // Use clientProtocol.NewDataPack

	headData := make([]byte, dp.GetHeadLen())
	if _, err := io.ReadFull(c.Conn, headData); err != nil {
		return nil, fmt.Errorf("read message head error: %w", err)
	}

	msg, err := dp.Unpack(headData) // msg is already *clientProtocol.Message from Unpack
	if err != nil {
		return nil, fmt.Errorf("unpack message head error: %w", err)
	}

	// msg := msgHead.(*clientProtocol.Message) // No type assertion needed here
	if msg.GetDataLen() > 0 {
		msg.Data = make([]byte, msg.GetDataLen())
		if _, err := io.ReadFull(c.Conn, msg.Data); err != nil {
			return nil, fmt.Errorf("read message data error: %w", err)
		}
	}
	return msg, nil
}

// SendHeartbeat 发送心跳消息
func (c *ChatClient) SendHeartbeat() error {
	return c.SendMessage(serverProtocol.MsgIDPing, []byte("ping"))
}

// StartHeartbeat 启动心跳
func (c *ChatClient) StartHeartbeat(interval time.Duration) {
	if c.heartbeatStop == nil {
		c.heartbeatStop = make(chan struct{})
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := c.SendHeartbeat(); err != nil {
					fmt.Printf("发送心跳失败: %v\n", err)
					return
				}
			case <-c.heartbeatStop:
				fmt.Println("心跳已停止.")
				return
			}
		}
	}()
}

// StopHeartbeat 停止心跳
func (c *ChatClient) StopHeartbeat() {
	if c.heartbeatStop != nil {
		select {
		case c.heartbeatStop <- struct{}{}:
		default:
		}
		close(c.heartbeatStop)
		c.heartbeatStop = nil
	}
}

// Register 注册用户
func (c *ChatClient) Register(username, password, email string) (*model.UserRegisterResponse, error) {
	req := model.UserRegisterReq{
		Username: username,
		Password: password,
		Email:    email,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal register request: %w", err)
	}
	if err := c.SendMessage(serverProtocol.MsgIDRegisterReq, body); err != nil {
		return nil, fmt.Errorf("发送注册请求失败: %v", err)
	}
	respMsg, err := c.readMessage()
	if err != nil {
		return nil, fmt.Errorf("读取注册响应失败: %v", err)
	}
	if respMsg.GetMsgID() != serverProtocol.MsgIDRegisterResp {
		return nil, fmt.Errorf("响应消息ID错误，期望%d，实际%d", serverProtocol.MsgIDRegisterResp, respMsg.GetMsgID())
	}
	var genericResp struct {
		Code uint32                     `json:"code"`
		Msg  string                     `json:"msg"`
		Data model.UserRegisterResponse `json:"data"`
	}
	if err := json.Unmarshal(respMsg.GetData(), &genericResp); err != nil {
		var mapResp map[string]interface{}
		if json.Unmarshal(respMsg.GetData(), &mapResp) == nil {
			fmt.Printf("Debug: Register Raw Response: %+v\n", mapResp)
		}
		return nil, fmt.Errorf("解析注册响应失败: %v, body: %s", err, string(respMsg.GetData()))
	}
	if genericResp.Code != 0 {
		return nil, fmt.Errorf("注册失败: %s (code: %d)", genericResp.Msg, genericResp.Code)
	}
	return &genericResp.Data, nil
}

// Login 用户登录
func (c *ChatClient) Login(username, password string) (*model.UserLoginResponse, error) {
	req := model.UserLoginReq{
		Username: username,
		Password: password,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal login request: %w", err)
	}
	if err := c.SendMessage(serverProtocol.MsgIDLoginReq, body); err != nil {
		return nil, fmt.Errorf("发送登录请求失败: %v", err)
	}
	respMsg, err := c.readMessage()
	if err != nil {
		return nil, fmt.Errorf("读取登录响应失败: %v", err)
	}
	if respMsg.GetMsgID() != serverProtocol.MsgIDLoginResp {
		return nil, fmt.Errorf("响应消息ID错误，期望%d，实际%d", serverProtocol.MsgIDLoginResp, respMsg.GetMsgID())
	}
	var genericResp struct {
		Code uint32                  `json:"code"`
		Msg  string                  `json:"msg"`
		Data model.UserLoginResponse `json:"data"`
	}
	if err := json.Unmarshal(respMsg.GetData(), &genericResp); err != nil {
		var mapResp map[string]interface{}
		if json.Unmarshal(respMsg.GetData(), &mapResp) == nil {
			fmt.Printf("Debug: Login Raw Response: %+v\n", mapResp)
		}
		return nil, fmt.Errorf("解析登录响应失败: %v, body: %s", err, string(respMsg.GetData()))
	}
	if genericResp.Code != 0 {
		return nil, fmt.Errorf("登录失败: %s (code: %d)", genericResp.Msg, genericResp.Code)
	}
	c.UserID = genericResp.Data.ID
	c.UserUUID = genericResp.Data.UserUUID
	c.Username = genericResp.Data.Username
	c.Token = genericResp.Data.Token
	c.isLoggedIn = true
	c.StartHeartbeat(60 * time.Second)
	fmt.Printf("用户 %s (ID: %d, UUID: %s) 登录成功.\n", c.Username, c.UserID, c.UserUUID)
	return &genericResp.Data, nil
}

// IsLoggedIn 检查客户端是否已登录
func (c *ChatClient) IsLoggedIn() bool {
	return c.isLoggedIn
}

// SendTextMessage 发送文本消息
func (c *ChatClient) SendTextMessage(toUserIdentity string, content string) error {
	if !c.isLoggedIn {
		return errors.New("请先登录再发送消息")
	}
	msg := model.TextMsg{
		ToUserID: toUserIdentity,
		Content:  content,
	}
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal text message: %w", err)
	}
	return c.SendMessage(serverProtocol.MsgIDTextMsg, body)
}

// StartMsgListener 启动消息监听器，接收服务器推送的消息
func (c *ChatClient) StartMsgListener(handler func(msgID uint32, data []byte)) {
	c.msgHandler = handler
	go func() {
		if c.Conn == nil {
			fmt.Println("错误：消息监听器启动时连接为空。")
			return
		}
		fmt.Println("消息监听器已启动...")
		for {
			msg, err := c.readMessage()
			if err != nil {
				if c.isLoggedIn {
					fmt.Printf("读取消息失败: %v. 监听器停止.\n", err)
				}
				c.Close()
				return
			}
			if c.msgHandler != nil {
				if msg.GetMsgID() == serverProtocol.MsgIDPong {
					continue
				}
				c.msgHandler(msg.GetMsgID(), msg.GetData())
			}
		}
	}()
}

// SendHistoryMessageReq 发送获取历史消息请求
func (c *ChatClient) SendHistoryMessageReq(targetUserUUID string, limit int) error {
	if !c.isLoggedIn {
		return errors.New("请先登录")
	}
	req := model.HistoryMsgReq{
		TargetUserUUID: targetUserUUID,
		Limit:          limit,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal history message request: %w", err)
	}
	return c.SendMessage(serverProtocol.MsgIDHistoryMsgReq, body)
}

// SendCreateGroupReq 发送创建群组请求
func (c *ChatClient) SendCreateGroupReq(name, description, avatar string) error {
	if !c.isLoggedIn {
		return errors.New("请先登录")
	}
	req := model.CreateGroupReq{
		Name:        name,
		Description: description,
		Avatar:      avatar,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal create group request: %w", err)
	}
	return c.SendMessage(serverProtocol.MsgIDCreateGroupReq, body)
}

// SendJoinGroupReq 发送加入群组请求
func (c *ChatClient) SendJoinGroupReq(groupID uint) error {
	if !c.isLoggedIn {
		return errors.New("请先登录")
	}
	req := model.JoinGroupReq{
		GroupID: groupID,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal join group request: %w", err)
	}
	return c.SendMessage(serverProtocol.MsgIDJoinGroupReq, body)
}

// SendLeaveGroupReq 发送离开群组请求
func (c *ChatClient) SendLeaveGroupReq(groupID uint) error {
	if !c.isLoggedIn {
		return errors.New("请先登录")
	}
	req := model.LeaveGroupReq{
		GroupID: groupID,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal leave group request: %w", err)
	}
	return c.SendMessage(serverProtocol.MsgIDLeaveGroupReq, body)
}
