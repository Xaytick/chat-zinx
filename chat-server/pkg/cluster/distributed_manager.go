package cluster

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/Xaytick/chat-zinx/chat-server/global"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/discovery"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/messaging"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
)

type DistributedManager struct {
	natsService   *messaging.NATSService
	consulService *discovery.ConsulService
	serverID      string
	serverAddr    string
	serverPort    int
	userSubs      map[string]bool
	groupSubs     map[string]bool
	mutex         sync.RWMutex
	isRunning     bool
}

func NewDistributedManager(serverID, serverAddr string, serverPort int) (*DistributedManager, error) {
	// 初始化NATS服务
	natsURLs := "nats://localhost:4222,nats://localhost:4223,nats://localhost:4224"
	if global.Config.IsDev {
		// 开发环境使用Docker容器名
		natsURLs = "nats://nats-1:4222,nats://nats-2:4222,nats://nats-3:4222"
	}

	natsService, err := messaging.NewNATSService(natsURLs, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to init NATS service: %w", err)
	}

	// 初始化Consul服务
	consulAddr := "localhost:8500"
	if global.Config.IsDev {
		consulAddr = "consul-1:8500"
	}

	consulConfig := &discovery.ServiceConfig{
		Name:    "chat-server",
		ID:      serverID,
		Address: serverAddr,
		Port:    serverPort,
		Tags:    []string{"chat", "realtime", "zinx"},
		Meta: map[string]string{
			"version":    "1.0.0",
			"region":     "default",
			"protocol":   "tcp",
			"framework":  "zinx",
			"start_time": fmt.Sprintf("%d", time.Now().Unix()),
		},
	}

	consulService, err := discovery.NewConsulService(consulAddr, consulConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to init Consul service: %w", err)
	}

	return &DistributedManager{
		natsService:   natsService,
		consulService: consulService,
		serverID:      serverID,
		serverAddr:    serverAddr,
		serverPort:    serverPort,
		userSubs:      make(map[string]bool),
		groupSubs:     make(map[string]bool),
		isRunning:     false,
	}, nil
}

// 启动分布式管理器
func (dm *DistributedManager) Start() error {
	// 注册服务到Consul
	if err := dm.consulService.RegisterService(); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	// 监听服务变更
	dm.consulService.WatchServices("chat-server", func(instances []string) {
		log.Printf("Chat server instances updated: %v", instances)
	})

	// 启动健康监控
	go dm.healthMonitor()

	dm.isRunning = true
	log.Printf("Distributed manager started for server %s", dm.serverID)

	// 发布服务启动事件
	dm.natsService.PublishSystemEvent("server_started", dm.serverID)

	return nil
}

// 处理用户上线
func (dm *DistributedManager) HandleUserOnline(userUUID string) error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	// 避免重复订阅
	if dm.userSubs[userUUID] {
		return nil
	}

	// 订阅用户个人消息
	err := dm.natsService.SubscribeUserMessages(userUUID, func(msg *messaging.CrossServerMessage) {
		dm.handleIncomingMessage(msg)
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe user messages: %w", err)
	}

	dm.userSubs[userUUID] = true

	// 在Consul中更新用户状态
	if err := dm.consulService.SetUserOnline(userUUID); err != nil {
		log.Printf("Failed to set user online in Consul: %v", err)
	}

	log.Printf("User %s online, subscribed to messages", userUUID)

	// 发布用户上线事件
	eventData := fmt.Sprintf(`{"user_uuid":"%s","server_id":"%s"}`, userUUID, dm.serverID)
	dm.natsService.PublishSystemEvent("user_online", eventData)

	return nil
}

// 处理用户下线
func (dm *DistributedManager) HandleUserOffline(userUUID string) error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	// 取消订阅
	if dm.userSubs[userUUID] {
		dm.natsService.Unsubscribe(userUUID)
		delete(dm.userSubs, userUUID)
	}

	// 从Consul删除用户状态
	if err := dm.consulService.SetUserOffline(userUUID); err != nil {
		log.Printf("Failed to set user offline in Consul: %v", err)
	}

	log.Printf("User %s offline, unsubscribed from messages", userUUID)

	// 发布用户下线事件
	eventData := fmt.Sprintf(`{"user_uuid":"%s","server_id":"%s"}`, userUUID, dm.serverID)
	dm.natsService.PublishSystemEvent("user_offline", eventData)

	return nil
}

// 发送跨服务器消息
func (dm *DistributedManager) SendCrossServerMessage(targetUserUUID string, message *model.TextMsg) error {
	// 检查目标用户是否在线
	targetServer, err := dm.consulService.GetUserServer(targetUserUUID)
	if err != nil {
		log.Printf("Target user %s not found online: %v", targetUserUUID, err)
		return err
	}

	// 如果用户在本服务器，直接返回
	if targetServer == dm.serverID {
		return fmt.Errorf("user is on local server")
	}

	log.Printf("Sending cross-server message to user %s on server %s", targetUserUUID, targetServer)
	return dm.natsService.SendP2PMessage(targetUserUUID, message)
}

// 发送群组消息
func (dm *DistributedManager) SendGroupMessage(groupID string, message *model.GroupTextMsgReq) error {
	log.Printf("Sending group message to group %s", groupID)
	return dm.natsService.SendGroupMessage(groupID, message)
}

// 订阅群组消息
func (dm *DistributedManager) SubscribeGroupMessages(groupID string) error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	key := fmt.Sprintf("group_%s", groupID)
	if dm.groupSubs[key] {
		return nil // 已订阅
	}

	err := dm.natsService.SubscribeGroupMessages(groupID, func(msg *messaging.CrossServerMessage) {
		dm.handleIncomingMessage(msg)
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe group messages: %w", err)
	}

	dm.groupSubs[key] = true
	log.Printf("Subscribed to group %s messages", groupID)
	return nil
}

// 处理接收到的消息
func (dm *DistributedManager) handleIncomingMessage(msg *messaging.CrossServerMessage) {
	switch msg.Type {
	case "p2p_message":
		dm.handleP2PMessage(msg)
	case "group_message":
		dm.handleGroupMessage(msg)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

func (dm *DistributedManager) handleP2PMessage(msg *messaging.CrossServerMessage) {
	// 将消息数据转换为TextMsg
	msgData, ok := msg.MessageData.(map[string]interface{})
	if !ok {
		log.Printf("Invalid P2P message data format")
		return
	}

	// 查找本地连接
	connManager := global.GlobalServer.GetConnManager()
	found := false

	for _, conn := range connManager.All() {
		if userUUIDProp, err := conn.GetProperty("userUUID"); err == nil {
			if userUUIDStr, ok := userUUIDProp.(string); ok && userUUIDStr == msg.TargetUserID {
				// 找到目标用户，转发消息
				jsonData, _ := json.Marshal(msgData)
				err := conn.SendMsg(protocol.MsgIDTextMsg, jsonData)
				if err != nil {
					log.Printf("Failed to forward P2P message: %v", err)
				} else {
					log.Printf("Successfully forwarded P2P message to user %s", msg.TargetUserID)
					found = true
				}
				break
			}
		}
	}

	if !found {
		log.Printf("Target user %s not found on this server", msg.TargetUserID)
	}
}

func (dm *DistributedManager) handleGroupMessage(msg *messaging.CrossServerMessage) {
	groupID := msg.TargetUserID

	// 将消息数据转换为GroupTextMsg
	msgData, ok := msg.MessageData.(map[string]interface{})
	if !ok {
		log.Printf("Invalid group message data format")
		return
	}

	// 获取本服务器上该群组的在线成员
	connManager := global.GlobalServer.GetConnManager()
	memberCount := 0

	for _, conn := range connManager.All() {
		// 检查用户是否是群组成员
		if userIDProp, err := conn.GetProperty("userID"); err == nil {
			userID, ok := userIDProp.(uint)
			if !ok {
				continue
			}

			// 检查用户是否在该群组中
			groupIDUint, err := strconv.ParseUint(groupID, 10, 32)
			if err != nil {
				log.Printf("Invalid group ID: %s", groupID)
				continue
			}

			isMember, err := global.GroupService.IsUserInGroup(userID, uint(groupIDUint))
			if err != nil || !isMember {
				continue
			}

			// 转发群组消息
			jsonData, _ := json.Marshal(msgData)
			err = conn.SendMsg(protocol.MsgIDGroupTextMsgResp, jsonData)
			if err != nil {
				log.Printf("Failed to forward group message: %v", err)
			} else {
				memberCount++
			}
		}
	}

	log.Printf("Forwarded group message to %d members on this server", memberCount)
}

// 获取集群状态
func (dm *DistributedManager) GetClusterStatus() map[string]interface{} {
	status := map[string]interface{}{
		"server_id":      dm.serverID,
		"is_running":     dm.isRunning,
		"nats_connected": dm.natsService.IsConnected(),
		"consul_healthy": dm.consulService.HealthCheck() == nil,
		"user_subs":      len(dm.userSubs),
		"group_subs":     len(dm.groupSubs),
		"timestamp":      time.Now().Unix(),
	}

	// 获取NATS统计信息
	if natsStats := dm.natsService.GetStats(); natsStats != nil {
		status["nats_stats"] = natsStats
	}

	// 获取Consul服务统计信息
	if consulStats, err := dm.consulService.GetServiceStats(); err == nil {
		status["consul_stats"] = consulStats
	}

	return status
}

// 获取在线用户分布
func (dm *DistributedManager) GetOnlineUserDistribution() (map[string]interface{}, error) {
	onlineUsers, err := dm.consulService.GetAllOnlineUsers()
	if err != nil {
		return nil, err
	}

	distribution := make(map[string]int)
	totalUsers := 0

	for _, serverID := range onlineUsers {
		distribution[serverID]++
		totalUsers++
	}

	result := map[string]interface{}{
		"total_users":  totalUsers,
		"distribution": distribution,
		"current_server": map[string]interface{}{
			"server_id":  dm.serverID,
			"user_count": distribution[dm.serverID],
			"local_subs": len(dm.userSubs),
		},
	}

	return result, nil
}

// 健康监控
func (dm *DistributedManager) healthMonitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !dm.isRunning {
				return
			}

			// 检查NATS连接
			if !dm.natsService.IsConnected() {
				log.Printf("NATS connection lost, attempting to reconnect...")
			}

			// 检查Consul连接
			if err := dm.consulService.HealthCheck(); err != nil {
				log.Printf("Consul health check failed: %v", err)
			}

			// 发布心跳事件
			heartbeat := map[string]interface{}{
				"server_id":  dm.serverID,
				"timestamp":  time.Now().Unix(),
				"user_count": len(dm.userSubs),
				"status":     "healthy",
			}
			if data, err := json.Marshal(heartbeat); err == nil {
				dm.natsService.PublishSystemEvent("heartbeat", string(data))
			}
		}
	}
}

// 广播系统消息
func (dm *DistributedManager) BroadcastSystemMessage(messageType, content string) error {
	event := map[string]interface{}{
		"type":      messageType,
		"content":   content,
		"server_id": dm.serverID,
		"timestamp": time.Now().Unix(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal system message: %w", err)
	}

	return dm.natsService.PublishSystemEvent("system_broadcast", string(data))
}

// 关闭分布式管理器
func (dm *DistributedManager) Stop() error {
	dm.isRunning = false

	// 发布服务停止事件
	dm.natsService.PublishSystemEvent("server_stopping", dm.serverID)

	// 清理所有用户订阅
	dm.mutex.Lock()
	for userUUID := range dm.userSubs {
		dm.consulService.SetUserOffline(userUUID)
	}
	dm.mutex.Unlock()

	// 注销服务
	if err := dm.consulService.DeregisterService(); err != nil {
		log.Printf("Failed to deregister service: %v", err)
	}

	// 关闭NATS连接
	if err := dm.natsService.Close(); err != nil {
		log.Printf("Failed to close NATS service: %v", err)
	}

	log.Printf("Distributed manager stopped for server %s", dm.serverID)
	return nil
}
