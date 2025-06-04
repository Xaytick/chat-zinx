# Chat-Zinx 分布式消息传递解决方案

## 🎯 技术选型分析

### 消息队列对比分析

| 特性               | **NATS** | **RabbitMQ** | **Kafka**   |
| ------------------ | -------------- | ------------------ | ----------------- |
| **延迟**     | 极低 (<1ms)    | 低 (1-5ms)         | 中等 (5-10ms)     |
| **吞吐量**   | 高 (1M+ msg/s) | 中等 (100K msg/s)  | 极高 (10M+ msg/s) |
| **复杂度**   | 简单           | 中等               | 复杂              |
| **内存使用** | 低             | 中等               | 高                |
| **持久化**   | 可选           | 支持               | 强制              |
| **集群管理** | 简单           | 中等               | 复杂              |
| **学习曲线** | 平缓           | 中等               | 陡峭              |
| **适用场景** | 实时通信       | 任务队列           | 流处理            |

### 服务发现对比分析

| 特性                 | **Consul** | **etcd** | **Nacos** |
| -------------------- | ---------------- | -------------- | --------------- |
| **性能**       | 高               | 极高           | 高              |
| **复杂度**     | 中等             | 低             | 中等            |
| **功能丰富度** | 极高             | 中等           | 高              |
| **社区活跃度** | 高               | 极高           | 中等            |
| **Go支持**     | 优秀             | 原生           | 良好            |

## 🏆 推荐方案：NATS + Consul

### 为什么选择这个组合？

#### NATS 优势：

- ✅ **超低延迟**: 微秒级消息传递，完美适配实时聊天
- ✅ **轻量级**: 单个二进制文件，部署简单
- ✅ **高性能**: 单节点百万消息/秒
- ✅ **原生Go支持**: 与项目技术栈完美匹配
- ✅ **Subject-based路由**: 天然支持用户/群组消息路由
- ✅ **集群简单**: 自动故障转移
- ✅ **JetStream**: 支持消息持久化和exactly-once语义

#### Consul 优势：

- ✅ **服务发现**: 自动注册和健康检查
- ✅ **配置管理**: 集中式配置存储
- ✅ **负载均衡**: 支持多种负载均衡策略
- ✅ **Multi-DC**: 支持多数据中心
- ✅ **Go原生客户端**: HashiCorp官方支持

## 🏗️ 完整架构设计

### 系统架构图

```
                    ┌─────────────────┐
                    │   Load Balancer │
                    │    (Nginx)      │
                    └─────────┬───────┘
                              │
                ┌─────────────┼─────────────┐
                │             │             │
        ┌───────▼────┐ ┌──────▼────┐ ┌──────▼────┐
        │Chat-Server1│ │Chat-Server2│ │Chat-Server3│
        │   :9000    │ │   :9001   │ │   :9002   │
        └─────┬──────┘ └─────┬─────┘ └─────┬─────┘
              │              │             │
              └──────────────┼─────────────┘
                             │
                    ┌────────▼────────┐
                    │  NATS Cluster   │
                    │ ┌─────┐ ┌─────┐ │
                    │ │NATS1│ │NATS2│ │
                    │ └─────┘ └─────┘ │
                    └─────────────────┘
                             │
                ┌────────────┼────────────┐
                │            │            │
        ┌───────▼──┐ ┌───────▼──┐ ┌───────▼──┐
        │ Consul1  │ │ Consul2  │ │ Consul3  │
        │ (Leader) │ │(Follower)│ │(Follower)│
        └──────────┘ └──────────┘ └──────────┘
                             │
                    ┌────────▼────────┐
                    │  Redis Cluster  │
                    │     (存储)      │
                    └─────────────────┘
```

### 消息流程设计

#### 1. 用户上线流程

```
用户连接Server1 → Server1注册到Consul → 用户状态写入Redis → 订阅NATS个人频道
```

#### 2. 跨服务器消息流程

```
用户A(Server1) → 发送消息 → 查询Redis找到用户B在Server2 → 
发送到NATS频道 → Server2接收 → 转发给用户B
```

#### 3. 群组消息流程

```
用户A发群消息 → 发送到NATS群组频道 → 所有Server接收 → 
各Server转发给本地群成员
```

## 🛠️ 具体实施方案

### 第一阶段：基础设施搭建

#### 1. Docker Compose 基础设施

```yaml
# infrastructure/docker-compose-messaging.yml
version: '3.8'

services:
  # NATS Cluster
  nats-1:
    image: nats:latest
    container_name: nats-1
    ports:
      - "4222:4222"
      - "8222:8222"
    command: 
      - "-cluster"
      - "nats://0.0.0.0:6222"
      - "-routes"
      - "nats-route://nats-2:6222,nats-route://nats-3:6222"
      - "-js"
      - "-sd"
      - "/data"
    volumes:
      - nats1_data:/data
    networks:
      - chat-network

  nats-2:
    image: nats:latest
    container_name: nats-2
    ports:
      - "4223:4222"
      - "8223:8222"
    command:
      - "-cluster"
      - "nats://0.0.0.0:6222"
      - "-routes"
      - "nats-route://nats-1:6222,nats-route://nats-3:6222"
      - "-js"
      - "-sd"
      - "/data"
    volumes:
      - nats2_data:/data
    networks:
      - chat-network

  nats-3:
    image: nats:latest
    container_name: nats-3
    ports:
      - "4224:4222"
      - "8224:8222"
    command:
      - "-cluster"
      - "nats://0.0.0.0:6222"
      - "-routes"
      - "nats-route://nats-1:6222,nats-route://nats-2:6222"
      - "-js"
      - "-sd"
      - "/data"
    volumes:
      - nats3_data:/data
    networks:
      - chat-network

  # Consul Cluster
  consul-1:
    image: hashicorp/consul:latest
    container_name: consul-1
    ports:
      - "8500:8500"
      - "8600:8600/udp"
    command:
      - "consul"
      - "agent"
      - "-server"
      - "-bootstrap-expect=3"
      - "-datacenter=dc1"
      - "-data-dir=/consul/data"
      - "-node=consul-1"
      - "-bind=0.0.0.0"
      - "-client=0.0.0.0"
      - "-retry-join=consul-2"
      - "-retry-join=consul-3"
      - "-ui"
    volumes:
      - consul1_data:/consul/data
    networks:
      - chat-network

  consul-2:
    image: hashicorp/consul:latest
    container_name: consul-2
    ports:
      - "8501:8500"
    command:
      - "consul"
      - "agent"
      - "-server"
      - "-bootstrap-expect=3"
      - "-datacenter=dc1"
      - "-data-dir=/consul/data"
      - "-node=consul-2"
      - "-bind=0.0.0.0"
      - "-client=0.0.0.0"
      - "-retry-join=consul-1"
      - "-retry-join=consul-3"
    volumes:
      - consul2_data:/consul/data
    networks:
      - chat-network

  consul-3:
    image: hashicorp/consul:latest
    container_name: consul-3
    ports:
      - "8502:8500"
    command:
      - "consul"
      - "agent"
      - "-server"
      - "-bootstrap-expect=3"
      - "-datacenter=dc1"
      - "-data-dir=/consul/data"
      - "-node=consul-3"
      - "-bind=0.0.0.0"
      - "-client=0.0.0.0"
      - "-retry-join=consul-1"
      - "-retry-join=consul-2"
    volumes:
      - consul3_data:/consul/data
    networks:
      - chat-network

volumes:
  nats1_data:
  nats2_data:
  nats3_data:
  consul1_data:
  consul2_data:
  consul3_data:

networks:
  chat-network:
    external: true
```

### 第二阶段：Go模块实现

#### 1. NATS消息服务

```go
// chat-server/pkg/messaging/nats_service.go
package messaging

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/nats-io/nats.go"
    "github.com/Xaytick/chat-zinx/chat-server/pkg/model"
)

type NATSService struct {
    nc       *nats.Conn
    js       nats.JetStreamContext
    serverID string
    subs     map[string]*nats.Subscription
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
    nc, err := nats.Connect(urls,
        nats.ReconnectWait(2*time.Second),
        nats.MaxReconnects(-1),
        nats.PingInterval(1*time.Minute),
        nats.MaxPingsOutstanding(2),
    )
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

    return service, nil
}

func (ns *NATSService) initStreams() error {
    // 创建聊天消息流
    _, err := ns.js.AddStream(&nats.StreamConfig{
        Name:     "CHAT_MESSAGES",
        Subjects: []string{"chat.>"},
        Storage:  nats.MemoryStorage,
        MaxAge:   24 * time.Hour,
    })
    if err != nil && err != nats.ErrStreamNameAlreadyInUse {
        return err
    }

    // 创建系统通知流
    _, err = ns.js.AddStream(&nats.StreamConfig{
        Name:     "SYSTEM_EVENTS",
        Subjects: []string{"system.>"},
        Storage:  nats.FileStorage,
        MaxAge:   7 * 24 * time.Hour,
    })
    if err != nil && err != nats.ErrStreamNameAlreadyInUse {
        return err
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
    return err
}

// 发送群组消息
func (ns *NATSService) SendGroupMessage(groupID string, message *model.GroupTextMsg) error {
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
        return fmt.Errorf("failed to marshal message: %w", err)
    }

    _, err = ns.js.Publish(subject, data)
    return err
}

// 订阅用户个人消息
func (ns *NATSService) SubscribeUserMessages(userID string, handler func(*CrossServerMessage)) error {
    subject := fmt.Sprintf("chat.p2p.%s", userID)
  
    sub, err := ns.js.Subscribe(subject, func(msg *nats.Msg) {
        var crossMsg CrossServerMessage
        if err := json.Unmarshal(msg.Data, &crossMsg); err != nil {
            log.Printf("Failed to unmarshal message: %v", err)
            return
        }
      
        // 避免处理自己发送的消息
        if crossMsg.SourceServer != ns.serverID {
            handler(&crossMsg)
        }
      
        msg.Ack()
    })
  
    if err != nil {
        return err
    }
  
    ns.subs[userID] = sub
    return nil
}

// 订阅群组消息
func (ns *NATSService) SubscribeGroupMessages(groupID string, handler func(*CrossServerMessage)) error {
    subject := fmt.Sprintf("chat.group.%s", groupID)
  
    sub, err := ns.js.Subscribe(subject, func(msg *nats.Msg) {
        var crossMsg CrossServerMessage
        if err := json.Unmarshal(msg.Data, &crossMsg); err != nil {
            log.Printf("Failed to unmarshal group message: %v", err)
            return
        }
      
        // 避免处理自己发送的消息
        if crossMsg.SourceServer != ns.serverID {
            handler(&crossMsg)
        }
      
        msg.Ack()
    })
  
    if err != nil {
        return err
    }
  
    ns.subs[fmt.Sprintf("group_%s", groupID)] = sub
    return nil
}

// 取消订阅
func (ns *NATSService) Unsubscribe(key string) error {
    if sub, exists := ns.subs[key]; exists {
        err := sub.Unsubscribe()
        delete(ns.subs, key)
        return err
    }
    return nil
}

// 关闭连接
func (ns *NATSService) Close() error {
    // 取消所有订阅
    for _, sub := range ns.subs {
        sub.Unsubscribe()
    }
  
    if ns.nc != nil {
        ns.nc.Close()
    }
    return nil
}
```

#### 2. Consul服务发现

```go
// chat-server/pkg/discovery/consul_service.go
package discovery

import (
    "fmt"
    "log"
    "net"
    "strconv"
    "time"

    "github.com/hashicorp/consul/api"
)

type ConsulService struct {
    client   *api.Client
    serverID string
    config   *ServiceConfig
}

type ServiceConfig struct {
    Name    string
    ID      string
    Address string
    Port    int
    Tags    []string
    Meta    map[string]string
}

func NewConsulService(consulAddr string, config *ServiceConfig) (*ConsulService, error) {
    consulConfig := api.DefaultConfig()
    consulConfig.Address = consulAddr
  
    client, err := api.NewClient(consulConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create consul client: %w", err)
    }

    return &ConsulService{
        client:   client,
        serverID: config.ID,
        config:   config,
    }, nil
}

// 注册服务
func (cs *ConsulService) RegisterService() error {
    // 健康检查配置
    check := &api.AgentServiceCheck{
        TCP:                            fmt.Sprintf("%s:%d", cs.config.Address, cs.config.Port),
        Interval:                       "10s",
        Timeout:                        "3s",
        DeregisterCriticalServiceAfter: "30s",
    }

    // 服务注册配置
    service := &api.AgentServiceRegistration{
        ID:      cs.config.ID,
        Name:    cs.config.Name,
        Tags:    cs.config.Tags,
        Address: cs.config.Address,
        Port:    cs.config.Port,
        Meta:    cs.config.Meta,
        Check:   check,
    }

    // 注册服务
    if err := cs.client.Agent().ServiceRegister(service); err != nil {
        return fmt.Errorf("failed to register service: %w", err)
    }

    log.Printf("Service %s registered successfully", cs.config.ID)
    return nil
}

// 注销服务
func (cs *ConsulService) DeregisterService() error {
    if err := cs.client.Agent().ServiceDeregister(cs.serverID); err != nil {
        return fmt.Errorf("failed to deregister service: %w", err)
    }
  
    log.Printf("Service %s deregistered successfully", cs.serverID)
    return nil
}

// 发现服务
func (cs *ConsulService) DiscoverServices(serviceName string) ([]*api.ServiceEntry, error) {
    services, _, err := cs.client.Health().Service(serviceName, "", true, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to discover services: %w", err)
    }
  
    return services, nil
}

// 获取健康的服务实例
func (cs *ConsulService) GetHealthyInstances(serviceName string) ([]string, error) {
    services, err := cs.DiscoverServices(serviceName)
    if err != nil {
        return nil, err
    }
  
    var instances []string
    for _, service := range services {
        instance := fmt.Sprintf("%s:%d", 
            service.Service.Address, 
            service.Service.Port)
        instances = append(instances, instance)
    }
  
    return instances, nil
}

// 监听服务变更
func (cs *ConsulService) WatchServices(serviceName string, callback func([]string)) error {
    plan, err := api.NewConnectCAQuery(api.QueryOptions{})
    if err != nil {
        return err
    }
  
    go func() {
        for {
            instances, err := cs.GetHealthyInstances(serviceName)
            if err != nil {
                log.Printf("Error getting healthy instances: %v", err)
                time.Sleep(5 * time.Second)
                continue
            }
          
            callback(instances)
            time.Sleep(10 * time.Second)
        }
    }()
  
    return nil
}

// 设置键值对
func (cs *ConsulService) SetKV(key, value string) error {
    kv := cs.client.KV()
    pair := &api.KVPair{
        Key:   key,
        Value: []byte(value),
    }
  
    _, err := kv.Put(pair, nil)
    return err
}

// 获取键值对
func (cs *ConsulService) GetKV(key string) (string, error) {
    kv := cs.client.KV()
    pair, _, err := kv.Get(key, nil)
    if err != nil {
        return "", err
    }
  
    if pair == nil {
        return "", fmt.Errorf("key not found")
    }
  
    return string(pair.Value), nil
}
```

#### 3. 集成到现有系统

```go
// chat-server/pkg/cluster/distributed_manager.go
package cluster

import (
    "fmt"
    "log"
    "strconv"
    "sync"

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
}

func NewDistributedManager(serverID, serverAddr string, serverPort int) (*DistributedManager, error) {
    // 初始化NATS服务
    natsURLs := "nats://nats-1:4222,nats://nats-2:4222,nats://nats-3:4222"
    natsService, err := messaging.NewNATSService(natsURLs, serverID)
    if err != nil {
        return nil, fmt.Errorf("failed to init NATS service: %w", err)
    }

    // 初始化Consul服务
    consulConfig := &discovery.ServiceConfig{
        Name:    "chat-server",
        ID:      serverID,
        Address: serverAddr,
        Port:    serverPort,
        Tags:    []string{"chat", "realtime"},
        Meta: map[string]string{
            "version": "1.0.0",
            "region":  "default",
        },
    }
  
    consulService, err := discovery.NewConsulService("consul-1:8500", consulConfig)
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
    }, nil
}

// 启动分布式管理器
func (dm *DistributedManager) Start() error {
    // 注册服务到Consul
    if err := dm.consulService.RegisterService(); err != nil {
        return fmt.Errorf("failed to register service: %w", err)
    }

    log.Printf("Distributed manager started for server %s", dm.serverID)
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
    key := fmt.Sprintf("users/online/%s", userUUID)
    value := fmt.Sprintf(`{"server_id":"%s","timestamp":%d}`, 
        dm.serverID, time.Now().Unix())
  
    return dm.consulService.SetKV(key, value)
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
    key := fmt.Sprintf("users/online/%s", userUUID)
    return dm.consulService.SetKV(key, "")
}

// 发送跨服务器消息
func (dm *DistributedManager) SendCrossServerMessage(targetUserUUID string, message *model.TextMsg) error {
    return dm.natsService.SendP2PMessage(targetUserUUID, message)
}

// 发送群组消息
func (dm *DistributedManager) SendGroupMessage(groupID string, message *model.GroupTextMsg) error {
    return dm.natsService.SendGroupMessage(groupID, message)
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
                }
                return
            }
        }
    }
  
    log.Printf("Target user %s not found on this server", msg.TargetUserID)
}

func (dm *DistributedManager) handleGroupMessage(msg *messaging.CrossServerMessage) {
    // 处理群组消息逻辑
    groupID := msg.TargetUserID
  
    // 获取本服务器上该群组的在线成员
    connManager := global.GlobalServer.GetConnManager()
    for _, conn := range connManager.All() {
        // 检查用户是否是群组成员
        if userIDProp, err := conn.GetProperty("userID"); err == nil {
            userID, ok := userIDProp.(uint)
            if !ok {
                continue
            }
          
            // 检查用户是否在该群组中
            groupIDUint, _ := strconv.ParseUint(groupID, 10, 32)
            isMember, err := global.GroupService.IsUserInGroup(userID, uint(groupIDUint))
            if err != nil || !isMember {
                continue
            }
          
            // 转发群组消息
            jsonData, _ := json.Marshal(msg.MessageData)
            err = conn.SendMsg(protocol.MsgIDGroupTextMsgResp, jsonData)
            if err != nil {
                log.Printf("Failed to forward group message: %v", err)
            }
        }
    }
}

// 关闭分布式管理器
func (dm *DistributedManager) Stop() error {
    // 注销服务
    if err := dm.consulService.DeregisterService(); err != nil {
        log.Printf("Failed to deregister service: %v", err)
    }

    // 关闭NATS连接
    if err := dm.natsService.Close(); err != nil {
        log.Printf("Failed to close NATS service: %v", err)
    }

    return nil
}
```

## 🚀 第三阶段：集成改造

### 修改主服务器启动逻辑

```go
// chat-server/main.go (增强版)
import (
    "github.com/Xaytick/chat-zinx/chat-server/pkg/cluster"
)

var distributedManager *cluster.DistributedManager

func main() {
    // ... 现有初始化逻辑 ...

    // 初始化分布式管理器
    serverID := fmt.Sprintf("chat-server-%s-%d", 
        config.Host, config.Port)
  
    var err error
    distributedManager, err = cluster.NewDistributedManager(
        serverID, config.Host, config.Port)
    if err != nil {
        log.Fatalf("Failed to create distributed manager: %v", err)
    }

    // 启动分布式服务
    if err := distributedManager.Start(); err != nil {
        log.Fatalf("Failed to start distributed manager: %v", err)
    }

    // 设置全局变量供其他模块使用
    global.DistributedManager = distributedManager

    // 设置连接开始时的钩子函数
    global.GlobalServer.SetOnConnStart(func(conn ziface.IConnection) {
        fmt.Println("新连接 ConnID=", conn.GetConnID(), "IP:", conn.RemoteAddr().String())
    })

    // 设置连接结束时的钩子函数
    global.GlobalServer.SetOnConnStop(func(conn ziface.IConnection) {
        if userUUIDProp, err := conn.GetProperty("userUUID"); err == nil {
            if userUUID, ok := userUUIDProp.(string); ok {
                // 处理用户下线
                distributedManager.HandleUserOffline(userUUID)
            }
        }
      
        if userID, err := conn.GetProperty("userID"); err == nil {
            username, _ := conn.GetProperty("username")
            fmt.Printf("连接断开 ConnID=%d, 用户: %s(ID=%v)\n",
                conn.GetConnID(), username, userID)
        }
    })

    // 设置优雅关闭
    defer func() {
        if distributedManager != nil {
            distributedManager.Stop()
        }
    }()

    // 启动服务器
    global.GlobalServer.Serve()
}
```

### 修改登录路由

```go
// chat-server/router/login.go (增强用户上线处理)
func (lr *LoginRouter) Handle(request ziface.IRequest) {
    // ... 现有登录逻辑 ...

    // 登录成功后的新增逻辑
    if global.DistributedManager != nil {
        // 处理用户上线，订阅个人消息
        err := global.DistributedManager.HandleUserOnline(user.UserUUID)
        if err != nil {
            fmt.Printf("[分布式] 处理用户上线失败: %v\n", err)
        } else {
            fmt.Printf("[分布式] 用户 %s 已加入分布式集群\n", user.Username)
        }
    }

    // ... 其余逻辑保持不变 ...
}
```

### 修改消息路由

```go
// chat-server/router/text.go (增强跨服务器支持)
func (r *TextMsgRouter) Handle(request ziface.IRequest) {
    // ... 现有解析和查找逻辑 ...

    // 先尝试本地投递
    foundOnline := false
    connManager := global.GlobalServer.GetConnManager()
  
    for _, conn := range connManager.All() {
        if userIDProp, err := conn.GetProperty("userID"); err == nil {
            userIDUint, ok := userIDProp.(uint)
            if ok && userIDUint == toUserIDUint {
                err := conn.SendMsg(protocol.MsgIDTextMsg, msgData)
                if err == nil {
                    foundOnline = true
                    break
                }
            }
        }
    }

    // 本地未找到，尝试跨服务器投递
    if !foundOnline && global.DistributedManager != nil {
        err := global.DistributedManager.SendCrossServerMessage(toUserUUIDStr, &msg)
        if err != nil {
            fmt.Printf("[跨服务器] 消息发送失败: %v\n", err)
        } else {
            fmt.Printf("[跨服务器] 消息已发送到分布式网络\n")
            foundOnline = true // 标记为已处理
        }
    }

    // 保存消息记录
    if foundOnline {
        global.MessageService.SaveHistoryOnly(fromUserIDUint, toUserIDUint, msgData)
    } else {
        global.MessageService.SaveMessage(fromUserIDUint, toUserIDUint, msgData)
    }
}
```

## 📊 性能与监控

### 监控指标

1. **NATS指标**: 消息吞吐量、延迟、队列深度
2. **Consul指标**: 服务健康状态、注册数量
3. **应用指标**: 跨服务器消息成功率、用户分布

### 性能优化

1. **连接池**: 复用NATS连接
2. **批量处理**: 合并小消息
3. **本地缓存**: 缓存服务发现结果
4. **负载均衡**: 基于连接数分配用户

## 🎯 部署与测试

### 部署步骤

1. 启动基础设施: `docker-compose -f infrastructure/docker-compose-messaging.yml up -d`
2. 启动多个聊天服务器实例
3. 验证服务注册和消息路由

### 测试场景

1. **单服务器**: 验证现有功能不受影响
2. **双服务器**: 测试跨服务器P2P消息
3. **多服务器**: 测试群组消息广播
4. **故障恢复**: 测试节点宕机恢复

这个解决方案将chat-zinx从单体架构升级为真正的分布式聊天系统，支持水平扩展和高可用性！
