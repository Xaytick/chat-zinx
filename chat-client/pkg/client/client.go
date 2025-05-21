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

	isLoggedIn       bool
	heartbeatStop    chan struct{}
	msgHandler       func(msgID uint32, data []byte)         // Callback for received messages
	responseChannels map[uint32]chan *clientProtocol.Message // New: Map to hold channels for pending responses
	requestTimeout   time.Duration                           // New: Timeout for requests
}

// NewChatClient 创建一个新的聊天客户端
func NewChatClient(serverAddr string) (*ChatClient, error) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("连接服务器失败: %v", err)
	}

	return &ChatClient{
		Conn:             conn,
		ServerAddr:       serverAddr,
		heartbeatStop:    make(chan struct{}),
		responseChannels: make(map[uint32]chan *clientProtocol.Message), // Initialize map
		requestTimeout:   10 * time.Second,                              // Default timeout
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

	// Create a channel for this specific request
	respChan := make(chan *clientProtocol.Message, 1)
	c.responseChannels[serverProtocol.MsgIDRegisterResp] = respChan // Using MsgID as key for simplicity here

	if err := c.SendMessage(serverProtocol.MsgIDRegisterReq, body); err != nil {
		delete(c.responseChannels, serverProtocol.MsgIDRegisterResp) // Clean up
		return nil, fmt.Errorf("发送注册请求失败: %v", err)
	}

	// Wait for the response on the channel or timeout
	select {
	case respMsg := <-respChan:
		if respMsg.GetMsgID() != serverProtocol.MsgIDRegisterResp { // Should be guaranteed by listener if key is correct
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
	case <-time.After(c.requestTimeout):
		delete(c.responseChannels, serverProtocol.MsgIDRegisterResp) // Clean up
		return nil, fmt.Errorf("注册响应超时")
	}
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

	respChan := make(chan *clientProtocol.Message, 1)
	c.responseChannels[serverProtocol.MsgIDLoginResp] = respChan

	if err := c.SendMessage(serverProtocol.MsgIDLoginReq, body); err != nil {
		delete(c.responseChannels, serverProtocol.MsgIDLoginResp)
		return nil, fmt.Errorf("发送登录请求失败: %v", err)
	}

	select {
	case respMsg := <-respChan:
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
		// Start heartbeat after successful login
		c.StartHeartbeat(30 * time.Second)
		return &genericResp.Data, nil
	case <-time.After(c.requestTimeout):
		delete(c.responseChannels, serverProtocol.MsgIDLoginResp)
		return nil, fmt.Errorf("登录响应超时")
	}
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

// StartMsgListener 启动消息监听器
func (c *ChatClient) StartMsgListener(handler func(msgID uint32, data []byte)) {
	c.msgHandler = handler
	go func() {
		for {
			msg, err := c.readMessage()
			if err != nil {
				if err == io.EOF || errors.Is(err, net.ErrClosed) {
					fmt.Printf("连接已关闭，停止监听消息: %v\n", err)
				} else {
					fmt.Printf("读取消息错误，停止监听: %v\n", err)
				}
				c.Close() // Ensure client is fully closed on read error
				return
			}

			// Check if this message ID is awaited by a synchronous call
			// Lock mechanism would be needed if responseChannels is accessed by multiple goroutines
			// For simplicity here, assuming StartMsgListener is the only writer to these channels.
			if ch, ok := c.responseChannels[msg.GetMsgID()]; ok {
				select {
				case ch <- msg:
					// Response sent to waiting synchronous call
				default:
					// Channel is full or not ready, could log this.
					// Or, if the design guarantees the channel is always ready, this case isn't needed.
					// For now, let's assume if a channel exists, it's ready.
					fmt.Printf("Warning: Response channel for MsgID %d was not ready or full.\n", msg.GetMsgID())
					// Fallback to general handler if channel send fails (e.g. full)
					if c.msgHandler != nil {
						c.msgHandler(msg.GetMsgID(), msg.GetData())
					}
				}
				// Once a response is routed, remove the channel to prevent leaks
				// and incorrect routing of future messages with the same ID.
				// This simple map keying by MsgID has limitations if multiple requests
				// expect the same response MsgID concurrently. A unique request ID would be better.
				delete(c.responseChannels, msg.GetMsgID())
			} else if c.msgHandler != nil {
				c.msgHandler(msg.GetMsgID(), msg.GetData())
			}
		}
	}()
}

// SendHistoryMessageReq 发送获取历史消息请求
func (c *ChatClient) SendHistoryMessageReq(targetUserIdentity string, limit int) error {
	if !c.isLoggedIn {
		return errors.New("请先登录")
	}
	req := model.LegacyHistoryMsgReq{
		// 客户端发送的是用户输入的标识，可能是Username，也可能是UUID
		// 根据服务器端的逻辑，它会先尝试UUID，再尝试Username
		TargetUserUUID: targetUserIdentity, // 先尝试作为UUID
		TargetUsername: targetUserIdentity, // 如果UUID不存在，则尝试Username
		Limit:          limit,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal history message request: %w", err)
	}

	// 历史消息请求通常期望一个响应，也应该使用 responseChannels 机制
	// 但当前的 HistoryMsgRouter 是直接推消息，没有让客户端等待特定响应的channel
	// 为了和 Register/Login 保持一致性，可以改造，但这里先保持原样，仅发送请求
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

// SendGroupTextMessage 发送群组文本消息
func (c *ChatClient) SendGroupTextMessage(groupID uint32, content string) error {
	if !c.isLoggedIn {
		return errors.New("请先登录再发送消息")
	}
	msg := model.GroupTextMsgReq{
		GroupID: groupID,
		Content: content,
	}
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal group text message: %w", err)
	}
	return c.SendMessage(serverProtocol.MsgIDGroupTextMsgReq, body)
}

// SendGroupHistoryMessageReq 发送获取群组历史消息请求
func (c *ChatClient) SendGroupHistoryMessageReq(groupID uint, lastID uint, limit int) error {
	if !c.isLoggedIn {
		return errors.New("请先登录")
	}
	req := model.GroupHistoryMsgReq{
		GroupID: groupID,
		LastID:  lastID,
		Limit:   limit,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal group history message request: %w", err)
	}
	return c.SendMessage(serverProtocol.MsgIDGroupHistoryMsgReq, body)
}
