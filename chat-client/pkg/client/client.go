package client

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
)

// ChatClient 聊天客户端结构体
type ChatClient struct {
	Conn     net.Conn
	UserID   string
	Username string
	// 用户名到用户ID的映射
	UsernameToID map[string]string
}

// NewChatClient 创建一个新的聊天客户端
func NewChatClient(serverAddr string) (*ChatClient, error) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("连接服务器失败: %v", err)
	}

	return &ChatClient{
		Conn:         conn,
		UsernameToID: make(map[string]string),
	}, nil
}

// Close 关闭客户端连接
func (c *ChatClient) Close() {
	if c.Conn != nil {
		c.Conn.Close()
	}
}

// SendMessage 发送消息
func (c *ChatClient) SendMessage(msgID uint32, body []byte) error {
	length := uint32(len(body))
	buf := make([]byte, 8+len(body))
	binary.LittleEndian.PutUint32(buf[0:4], length)
	binary.LittleEndian.PutUint32(buf[4:8], msgID)
	copy(buf[8:], body)
	_, err := c.Conn.Write(buf)
	return err
}

// ReadResponse 读取响应
func (c *ChatClient) ReadResponse() ([]byte, []byte, error) {
	head := make([]byte, 8)
	_, err := c.Conn.Read(head)
	if err != nil {
		return nil, nil, fmt.Errorf("读取头部失败: %v", err)
	}

	respLen := binary.LittleEndian.Uint32(head[0:4])
	respBody := make([]byte, respLen)
	_, err = c.Conn.Read(respBody)
	if err != nil {
		return nil, nil, fmt.Errorf("读取消息体失败: %v", err)
	}

	return head, respBody, nil
}

// Register 注册用户
func (c *ChatClient) Register(username, password, email string) (map[string]interface{}, error) {
	req := map[string]string{
		"username": username,
		"password": password,
		"email":    email,
	}
	body, _ := json.Marshal(req)

	if err := c.SendMessage(protocol.MsgIDRegisterReq, body); err != nil {
		return nil, fmt.Errorf("发送注册请求失败: %v", err)
	}

	head, respBody, err := c.ReadResponse()
	if err != nil {
		return nil, err
	}

	msgID := binary.LittleEndian.Uint32(head[4:8])
	if msgID != protocol.MsgIDRegisterResp {
		return nil, fmt.Errorf("响应消息ID错误，期望%d，实际%d", protocol.MsgIDRegisterResp, msgID)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 如果注册成功，保存用户ID
	if resp["code"].(float64) == 0 && resp["data"] != nil {
		data := resp["data"].(map[string]interface{})
		if userID, ok := data["user_id"].(string); ok && username != "" {
			c.UsernameToID[username] = userID
		}
	}

	return resp, nil
}

// Login 用户登录
func (c *ChatClient) Login(username, password string) (map[string]interface{}, error) {
	req := map[string]string{
		"username": username,
		"password": password,
	}
	body, _ := json.Marshal(req)

	if err := c.SendMessage(protocol.MsgIDLoginReq, body); err != nil {
		return nil, fmt.Errorf("发送登录请求失败: %v", err)
	}

	head, respBody, err := c.ReadResponse()
	if err != nil {
		return nil, err
	}

	msgID := binary.LittleEndian.Uint32(head[4:8])
	if msgID != protocol.MsgIDLoginResp {
		return nil, fmt.Errorf("响应消息ID错误，期望%d，实际%d", protocol.MsgIDLoginResp, msgID)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 如果登录成功，设置客户端的用户信息
	if resp["code"].(float64) == 0 {
		data := resp["data"].(map[string]interface{})
		c.UserID = data["user_id"].(string)
		c.Username = data["username"].(string)

		// 同时保存到映射表中
		c.UsernameToID[c.Username] = c.UserID
	}

	return resp, nil
}

// GetUserID 根据用户名获取用户ID
func (c *ChatClient) GetUserID(username string) string {
	// 如果是自己，直接返回自己的ID
	if username == c.Username {
		return c.UserID
	}

	// 从映射表中查找
	if id, ok := c.UsernameToID[username]; ok {
		return id
	}

	// 如果没有找到，默认返回用户名（向后兼容）
	return username
}

// AddUserInfo 添加用户信息到映射表
func (c *ChatClient) AddUserInfo(username, userID string) {
	if username != "" && userID != "" {
		c.UsernameToID[username] = userID
		fmt.Printf("添加用户映射: %s -> %s\n", username, userID)
	}
}

// SendTextMessage 发送文本消息
func (c *ChatClient) SendTextMessage(toUsername, content string) error {
	// 获取接收者的用户ID
	toUserID := c.GetUserID(toUsername)

	msg := map[string]string{
		"to_user_id": toUserID,
		"content":    content,
	}
	body, _ := json.Marshal(msg)
	return c.SendMessage(protocol.MsgIDTextMsg, body)
}

// StartMsgListener 启动消息监听器，接收服务器推送的消息
func (c *ChatClient) StartMsgListener(handler func(msgID uint32, msgBody []byte)) {
	go func() {
		for {
			head, body, err := c.ReadResponse()
			if err != nil {
				fmt.Println("连接关闭或读取出错:", err)
				return
			}

			msgID := binary.LittleEndian.Uint32(head[4:8])
			// 首先尝试格式化显示消息内容
			if msgID == protocol.MsgIDTextMsg {
				var textMsg model.TextMsg
				if err := json.Unmarshal(body, &textMsg); err == nil {
					fmt.Printf("\n[接收消息] 来自: %s, 内容: %s\n",
						c.GetUsernameByID(textMsg.ToUserID), textMsg.Content)
				}
			}

			// 调用自定义处理函数
			handler(msgID, body)
		}
	}()
}

// GetUsernameByID 根据用户ID获取用户名
func (c *ChatClient) GetUsernameByID(userID string) string {
	// 如果是自己的ID，返回自己的用户名
	if userID == c.UserID {
		return c.Username
	}

	// 遍历映射表查找
	for username, id := range c.UsernameToID {
		if id == userID {
			return username
		}
	}

	// 如果找不到，返回原始ID
	return userID
}

// RegisterAndLogin 注册并登录（如果用户已存在则直接登录）
func (c *ChatClient) RegisterAndLogin(username, password string) error {
	// 1. 尝试注册
	fmt.Println("尝试注册用户:", username)
	resp, err := c.Register(username, password, username+"@example.com")
	if err != nil {
		return err
	}

	fmt.Println("注册响应:", resp["msg"])

	// 等待一会，避免消息发送太快
	time.Sleep(500 * time.Millisecond)

	// 2. 尝试登录
	fmt.Println("尝试登录用户:", username)
	loginResp, err := c.Login(username, password)
	if err != nil {
		return err
	}

	if loginResp["code"].(float64) != 0 {
		return fmt.Errorf("登录失败: %s", loginResp["msg"].(string))
	}

	fmt.Println("登录成功! 用户信息:", loginResp["data"])
	return nil
}
