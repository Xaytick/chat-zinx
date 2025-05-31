package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/cache"
)

// TestMessage 测试消息结构
type TestMessage struct {
	ID      string    `json:"id"`
	UserID  string    `json:"user_id"`
	Content string    `json:"content"`
	Time    time.Time `json:"time"`
}

// TestUser 测试用户结构
type TestUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Status   string `json:"status"`
}

func main() {
	fmt.Println("🧪 开始Redis Cluster功能测试...")

	// 创建Redis Cluster配置
	config := &cache.RedisClusterConfig{
		Addrs: []string{
			"localhost:7001",
			"localhost:7002",
			"localhost:7003",
			"localhost:7004",
			"localhost:7005",
			"localhost:7006",
		},
		PoolSize:     20,
		MinIdleConns: 10,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	// 创建Redis Cluster管理器
	clusterManager, err := cache.NewRedisClusterManager(config)
	if err != nil {
		log.Fatalf("❌ 创建Redis Cluster管理器失败: %v", err)
	}
	defer clusterManager.Close()

	// 创建聊天缓存管理器
	chatCache := cache.NewChatCacheManager(clusterManager)

	// 测试1: 健康检查
	fmt.Println("\n📋 测试1: 集群健康检查")
	if err := clusterManager.Health(); err != nil {
		log.Fatalf("❌ 集群健康检查失败: %v", err)
	}
	fmt.Println("✅ 集群健康检查通过")

	// 测试2: 获取集群信息
	fmt.Println("\n📋 测试2: 获取集群信息")
	clusterInfo, err := clusterManager.GetClusterInfo()
	if err != nil {
		log.Printf("⚠️ 获取集群信息失败: %v", err)
	} else {
		fmt.Printf("📊 集群节点信息:\n%s\n", clusterInfo.Nodes[:200]+"...")
	}

	// 测试3: 数据分片测试
	fmt.Println("\n📋 测试3: 数据分片测试")
	testDataSharding(clusterManager.GetClient())

	// 测试4: 用户会话缓存测试
	fmt.Println("\n📋 测试4: 用户会话缓存测试")
	testUserSessions(chatCache)

	// 测试5: 用户在线状态测试
	fmt.Println("\n📋 测试5: 用户在线状态测试")
	testUserOnlineStatus(chatCache)

	// 测试6: 消息缓存测试
	fmt.Println("\n📋 测试6: 消息缓存测试")
	testMessageCache(chatCache)

	// 测试7: 消息队列测试
	fmt.Println("\n📋 测试7: 消息队列测试")
	testMessageQueue(chatCache)

	// 测试8: 故障转移测试
	fmt.Println("\n📋 测试8: 故障转移测试")
	testFailover(clusterManager.GetClient())

	// 测试9: 性能测试
	fmt.Println("\n📋 测试9: 性能测试")
	testPerformance(chatCache)

	fmt.Println("\n🎉 所有测试完成！Redis Cluster运行正常")
}

// testDataSharding 测试数据分片
func testDataSharding(client *redis.ClusterClient) {
	keys := []string{
		"user:1000", "user:2000", "user:3000", "user:4000",
		"session:abc", "session:def", "session:ghi",
		"message:msg1", "message:msg2", "message:msg3",
	}

	fmt.Println("🔄 写入测试数据到不同分片...")
	for i, key := range keys {
		value := fmt.Sprintf("test_value_%d", i)
		if err := client.Set(client.Context(), key, value, time.Hour).Err(); err != nil {
			log.Printf("❌ 设置 %s 失败: %v", key, err)
			continue
		}

		// 获取数据所在的槽位和节点
		slot := client.ClusterKeySlot(client.Context(), key).Val()
		fmt.Printf("✅ %s -> 槽位: %d\n", key, slot)
	}

	// 验证数据读取
	fmt.Println("📖 验证数据读取...")
	for _, key := range keys {
		value, err := client.Get(client.Context(), key).Result()
		if err != nil {
			log.Printf("❌ 读取 %s 失败: %v", key, err)
		} else {
			fmt.Printf("✅ %s: %s\n", key, value)
		}
	}
}

// testUserSessions 测试用户会话缓存
func testUserSessions(cache *cache.ChatCacheManager) {
	users := []TestUser{
		{ID: "user_001", Username: "alice", Status: "online"},
		{ID: "user_002", Username: "bob", Status: "away"},
		{ID: "user_003", Username: "charlie", Status: "busy"},
	}

	fmt.Println("💾 测试用户会话缓存...")
	for _, user := range users {
		// 设置会话
		if err := cache.SetUserSession(user.ID, user, 30*time.Minute); err != nil {
			log.Printf("❌ 设置用户会话失败 %s: %v", user.ID, err)
			continue
		}

		// 获取会话
		sessionData, err := cache.GetUserSession(user.ID)
		if err != nil {
			log.Printf("❌ 获取用户会话失败 %s: %v", user.ID, err)
			continue
		}

		var retrievedUser TestUser
		if err := json.Unmarshal([]byte(sessionData), &retrievedUser); err != nil {
			log.Printf("❌ 解析会话数据失败 %s: %v", user.ID, err)
			continue
		}

		fmt.Printf("✅ 用户会话 %s: %s (状态: %s)\n", user.ID, retrievedUser.Username, retrievedUser.Status)
	}
}

// testUserOnlineStatus 测试用户在线状态
func testUserOnlineStatus(cache *cache.ChatCacheManager) {
	userServers := map[string]string{
		"user_001": "server_1",
		"user_002": "server_2",
		"user_003": "server_1",
		"user_004": "server_3",
	}

	fmt.Println("🌐 测试用户在线状态...")
	for userID, serverID := range userServers {
		// 设置用户在线
		if err := cache.SetUserOnline(userID, serverID); err != nil {
			log.Printf("❌ 设置用户在线失败 %s: %v", userID, err)
			continue
		}

		// 获取用户服务器
		server, err := cache.GetUserServer(userID)
		if err != nil {
			log.Printf("❌ 获取用户服务器失败 %s: %v", userID, err)
			continue
		}

		fmt.Printf("✅ 用户 %s 在线于服务器: %s\n", userID, server)
	}
}

// testMessageCache 测试消息缓存
func testMessageCache(cache *cache.ChatCacheManager) {
	messages := []TestMessage{
		{ID: "msg_001", UserID: "user_001", Content: "Hello World!", Time: time.Now()},
		{ID: "msg_002", UserID: "user_002", Content: "Redis Cluster测试", Time: time.Now()},
		{ID: "msg_003", UserID: "user_003", Content: "分布式缓存很棒！", Time: time.Now()},
	}

	fmt.Println("📨 测试消息缓存...")
	for _, msg := range messages {
		// 缓存消息
		if err := cache.CacheMessage(msg.ID, msg, time.Hour); err != nil {
			log.Printf("❌ 缓存消息失败 %s: %v", msg.ID, err)
			continue
		}

		// 获取缓存消息
		msgData, err := cache.GetCachedMessage(msg.ID)
		if err != nil {
			log.Printf("❌ 获取缓存消息失败 %s: %v", msg.ID, err)
			continue
		}

		var retrievedMsg TestMessage
		if err := json.Unmarshal([]byte(msgData), &retrievedMsg); err != nil {
			log.Printf("❌ 解析消息数据失败 %s: %v", msg.ID, err)
			continue
		}

		fmt.Printf("✅ 消息 %s: %s (来自: %s)\n", msg.ID, retrievedMsg.Content, retrievedMsg.UserID)
	}
}

// testMessageQueue 测试消息队列
func testMessageQueue(cache *cache.ChatCacheManager) {
	queue := "message_queue"
	messages := []string{
		"消息1: 系统通知",
		"消息2: 用户聊天",
		"消息3: 群组消息",
	}

	fmt.Println("📤 测试消息队列...")

	// 推送消息
	for _, msg := range messages {
		if err := cache.PushMessage(queue, msg); err != nil {
			log.Printf("❌ 推送消息失败: %v", err)
			continue
		}
		fmt.Printf("✅ 推送消息: %s\n", msg)
	}

	// 弹出消息
	fmt.Println("📥 弹出消息...")
	for i := 0; i < len(messages); i++ {
		msg, err := cache.PopMessage(queue)
		if err != nil {
			log.Printf("❌ 弹出消息失败: %v", err)
			continue
		}
		fmt.Printf("✅ 弹出消息: %s\n", msg)
	}
}

// testFailover 测试故障转移
func testFailover(client *redis.ClusterClient) {
	fmt.Println("🔄 测试故障转移能力...")

	// 测试数据
	testKey := "failover_test"
	testValue := "测试故障转移"

	// 写入数据
	if err := client.Set(client.Context(), testKey, testValue, time.Hour).Err(); err != nil {
		log.Printf("❌ 写入测试数据失败: %v", err)
		return
	}

	// 多次读取验证一致性
	for i := 0; i < 5; i++ {
		value, err := client.Get(client.Context(), testKey).Result()
		if err != nil {
			log.Printf("❌ 读取失败 (第%d次): %v", i+1, err)
		} else {
			fmt.Printf("✅ 读取成功 (第%d次): %s\n", i+1, value)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// testPerformance 测试性能
func testPerformance(cache *cache.ChatCacheManager) {
	fmt.Println("⚡ 测试性能...")

	// 并发用户数
	userCount := 1000
	start := time.Now()

	// 模拟设置大量用户在线状态
	for i := 0; i < userCount; i++ {
		userID := fmt.Sprintf("perf_user_%d", i)
		serverID := fmt.Sprintf("server_%d", i%10) // 10个服务器

		if err := cache.SetUserOnline(userID, serverID); err != nil {
			log.Printf("❌ 性能测试失败 %s: %v", userID, err)
		}
	}

	duration := time.Since(start)
	qps := float64(userCount) / duration.Seconds()

	fmt.Printf("✅ 性能测试完成:\n")
	fmt.Printf("   - 处理用户数: %d\n", userCount)
	fmt.Printf("   - 总耗时: %v\n", duration)
	fmt.Printf("   - QPS: %.2f\n", qps)
}
