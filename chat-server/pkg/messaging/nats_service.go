package messaging

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"github.com/nats-io/nats.go"
)

type NATSService struct {
	nc       *nats.Conn
	js       nats.JetStreamContext
	serverID string
	subs     map[string]*nats.Subscription
	mutex    sync.RWMutex
}

type CrossServerMessage struct {
	Type         string      `json:"type"`
	TargetUserID string      `json:"target_user_id"`
	SourceServer string      `json:"source_server"`
	MessageData  interface{} `json:"message_data"`
	Timestamp    time.Time   `json:"timestamp"`
}

func NewNATSService(urls, serverID string) (*NATSService, error) {
	// 连接到NATS集群
	opts := []nats.Option{
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(-1),
		nats.PingInterval(1 * time.Minute),
		nats.MaxPingsOutstanding(2),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("NATS reconnected to %v", nc.ConnectedUrl())
		}),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Printf("NATS disconnected: %v", err)
		}),
	}

	nc, err := nats.Connect(urls, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// 创建JetStream上下文
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	service := &NATSService{
		nc:       nc,
		js:       js,
		serverID: serverID,
		subs:     make(map[string]*nats.Subscription),
	}

	// 初始化流
	if err := service.initStreams(); err != nil {
		return nil, fmt.Errorf("failed to init streams: %w", err)
	}

	log.Printf("NATS service initialized for server %s", serverID)
	return service, nil
}

func (ns *NATSService) initStreams() error {
	// 创建聊天消息流
	streamConfig := &nats.StreamConfig{
		Name:     "CHAT_MESSAGES",
		Subjects: []string{"chat.>"},
		Storage:  nats.MemoryStorage,
		MaxAge:   24 * time.Hour,
		MaxMsgs:  1000000,
		Replicas: 3, // 3个副本保证高可用
	}

	_, err := ns.js.AddStream(streamConfig)
	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		log.Printf("Warning: failed to create CHAT_MESSAGES stream: %v", err)
	}

	// 创建系统通知流
	systemStreamConfig := &nats.StreamConfig{
		Name:     "SYSTEM_EVENTS",
		Subjects: []string{"system.>"},
		Storage:  nats.FileStorage,
		MaxAge:   7 * 24 * time.Hour,
		Replicas: 3,
	}

	_, err = ns.js.AddStream(systemStreamConfig)
	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		log.Printf("Warning: failed to create SYSTEM_EVENTS stream: %v", err)
	}

	return nil
}

// 发送点对点消息
func (ns *NATSService) SendP2PMessage(targetUserID string, message *model.TextMsg) error {
	subject := fmt.Sprintf("chat.p2p.%s", targetUserID)

	crossMsg := CrossServerMessage{
		Type:         "p2p_message",
		TargetUserID: targetUserID,
		SourceServer: ns.serverID,
		MessageData:  message,
		Timestamp:    time.Now(),
	}

	data, err := json.Marshal(crossMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = ns.js.Publish(subject, data)
	if err != nil {
		log.Printf("Failed to publish P2P message to %s: %v", targetUserID, err)
		return err
	}

	log.Printf("P2P message sent to %s via NATS", targetUserID)
	return nil
}

// 发送群组消息
func (ns *NATSService) SendGroupMessage(groupID string, message *model.GroupTextMsgReq) error {
	subject := fmt.Sprintf("chat.group.%s", groupID)

	crossMsg := CrossServerMessage{
		Type:         "group_message",
		TargetUserID: groupID,
		SourceServer: ns.serverID,
		MessageData:  message,
		Timestamp:    time.Now(),
	}

	data, err := json.Marshal(crossMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal group message: %w", err)
	}

	_, err = ns.js.Publish(subject, data)
	if err != nil {
		log.Printf("Failed to publish group message to %s: %v", groupID, err)
		return err
	}

	log.Printf("Group message sent to group %s via NATS", groupID)
	return nil
}

// 订阅用户个人消息
func (ns *NATSService) SubscribeUserMessages(userID string, handler func(*CrossServerMessage)) error {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()

	subject := fmt.Sprintf("chat.p2p.%s", userID)

	// 避免重复订阅
	if _, exists := ns.subs[userID]; exists {
		return nil
	}

	sub, err := ns.js.Subscribe(subject, func(msg *nats.Msg) {
		var crossMsg CrossServerMessage
		if err := json.Unmarshal(msg.Data, &crossMsg); err != nil {
			log.Printf("Failed to unmarshal P2P message: %v", err)
			msg.Nak()
			return
		}

		// 避免处理自己发送的消息
		if crossMsg.SourceServer != ns.serverID {
			handler(&crossMsg)
		}

		msg.Ack()
	}, nats.Durable(fmt.Sprintf("user_%s_%s", userID, ns.serverID)))

	if err != nil {
		return fmt.Errorf("failed to subscribe to user messages: %w", err)
	}

	ns.subs[userID] = sub
	log.Printf("Subscribed to messages for user %s", userID)
	return nil
}

// 订阅群组消息
func (ns *NATSService) SubscribeGroupMessages(groupID string, handler func(*CrossServerMessage)) error {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()

	subject := fmt.Sprintf("chat.group.%s", groupID)
	key := fmt.Sprintf("group_%s", groupID)

	// 避免重复订阅
	if _, exists := ns.subs[key]; exists {
		return nil
	}

	sub, err := ns.js.Subscribe(subject, func(msg *nats.Msg) {
		var crossMsg CrossServerMessage
		if err := json.Unmarshal(msg.Data, &crossMsg); err != nil {
			log.Printf("Failed to unmarshal group message: %v", err)
			msg.Nak()
			return
		}

		// 避免处理自己发送的消息
		if crossMsg.SourceServer != ns.serverID {
			handler(&crossMsg)
		}

		msg.Ack()
	}, nats.Durable(fmt.Sprintf("group_%s_%s", groupID, ns.serverID)))

	if err != nil {
		return fmt.Errorf("failed to subscribe to group messages: %w", err)
	}

	ns.subs[key] = sub
	log.Printf("Subscribed to messages for group %s", groupID)
	return nil
}

// 取消订阅
func (ns *NATSService) Unsubscribe(key string) error {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()

	if sub, exists := ns.subs[key]; exists {
		err := sub.Unsubscribe()
		delete(ns.subs, key)
		log.Printf("Unsubscribed from %s", key)
		return err
	}
	return nil
}

// 发布系统事件
func (ns *NATSService) PublishSystemEvent(eventType, data string) error {
	subject := fmt.Sprintf("system.%s", eventType)

	event := map[string]interface{}{
		"server_id": ns.serverID,
		"type":      eventType,
		"data":      data,
		"timestamp": time.Now(),
	}

	jsonData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal system event: %w", err)
	}

	_, err = ns.js.Publish(subject, jsonData)
	return err
}

// 获取连接状态
func (ns *NATSService) IsConnected() bool {
	return ns.nc != nil && ns.nc.IsConnected()
}

// 获取统计信息
func (ns *NATSService) GetStats() map[string]interface{} {
	if ns.nc == nil {
		return nil
	}

	stats := ns.nc.Stats()
	return map[string]interface{}{
		"connected":     ns.nc.IsConnected(),
		"server_id":     ns.serverID,
		"in_msgs":       stats.InMsgs,
		"out_msgs":      stats.OutMsgs,
		"in_bytes":      stats.InBytes,
		"out_bytes":     stats.OutBytes,
		"reconnects":    stats.Reconnects,
		"subscriptions": len(ns.subs),
	}
}

// 关闭连接
func (ns *NATSService) Close() error {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()

	// 取消所有订阅
	for key, sub := range ns.subs {
		if err := sub.Unsubscribe(); err != nil {
			log.Printf("Error unsubscribing from %s: %v", key, err)
		}
	}
	ns.subs = make(map[string]*nats.Subscription)

	if ns.nc != nil {
		ns.nc.Close()
		log.Printf("NATS service closed for server %s", ns.serverID)
	}
	return nil
}
