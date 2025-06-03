package main

import (
	"database/sql"
	"fmt"
	"net"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	fmt.Println("🚀 Chat-Zinx 完整系统测试")
	fmt.Println(strings.Repeat("=", 80))

	// 1. 测试网络连接
	fmt.Println("\n📡 测试网络连接...")
	testNetworkConnections()

	// 2. 测试数据库连接
	fmt.Println("\n💾 测试数据库连接...")
	testDatabaseConnections()

	// 3. 测试主从复制
	fmt.Println("\n🔄 测试主从复制...")
	testMasterSlaveReplication()

	// 4. 测试Redis连接
	fmt.Println("\n🗄️ 测试Redis连接...")
	testRedisConnection()

	// 5. 总结报告
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("🎉 系统测试完成!")
	fmt.Println("\n📊 服务访问地址:")
	fmt.Println("   🌐 Chat Server (TCP): localhost:9000")
	fmt.Println("   🌐 Chat Server (HTTP): localhost:8080")
	fmt.Println("   🔧 Adminer (数据库管理): http://localhost:8081")
	fmt.Println("   📈 Grafana (监控): http://localhost:3000 (admin/admin)")
	fmt.Println("   📊 Prometheus: http://localhost:9090")
	fmt.Println("\n💡 测试建议:")
	fmt.Println("   1. 使用chat客户端连接到 localhost:9000")
	fmt.Println("   2. 通过Adminer查看数据库数据同步情况")
	fmt.Println("   3. 创建用户和发送消息测试分片功能")
}

func testNetworkConnections() {
	services := map[string]string{
		"Chat Server (TCP)":  "localhost:9000",
		"Chat Server (HTTP)": "localhost:8080",
		"Adminer":            "localhost:8081",
		"MySQL Master":       "localhost:3316",
		"MySQL Slave1":       "localhost:3317",
		"MySQL Slave2":       "localhost:3318",
		"MySQL Shard0":       "localhost:3320",
		"MySQL Shard1":       "localhost:3321",
		"Redis":              "localhost:6379",
	}

	for name, addr := range services {
		conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
		if err != nil {
			fmt.Printf("   ❌ %s (%s): %v\n", name, addr, err)
		} else {
			fmt.Printf("   ✅ %s (%s): 连接成功\n", name, addr)
			conn.Close()
		}
	}
}

func testDatabaseConnections() {
	databases := map[string]string{
		"主库":  "chatuser:chatpassword@tcp(localhost:3316)/chat_app",
		"从库1": "chatuser:chatpassword@tcp(localhost:3317)/chat_app",
		"从库2": "chatuser:chatpassword@tcp(localhost:3318)/chat_app",
		"分片0": "chatuser:chatpassword@tcp(localhost:3320)/chat_app_shard_00",
		"分片1": "chatuser:chatpassword@tcp(localhost:3321)/chat_app_shard_01",
	}

	for name, dsn := range databases {
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			fmt.Printf("   ❌ %s: 连接失败 - %v\n", name, err)
			continue
		}
		defer db.Close()

		err = db.Ping()
		if err != nil {
			fmt.Printf("   ❌ %s: Ping失败 - %v\n", name, err)
			continue
		}

		// 查询表数量
		var tableCount int
		var schema string
		if name == "分片0" {
			schema = "chat_app_shard_00"
		} else if name == "分片1" {
			schema = "chat_app_shard_01"
		} else {
			schema = "chat_app"
		}

		query := "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ?"
		err = db.QueryRow(query, schema).Scan(&tableCount)
		if err != nil {
			fmt.Printf("   ❌ %s: 查询表失败 - %v\n", name, err)
			continue
		}

		fmt.Printf("   ✅ %s: 连接成功，表数量: %d\n", name, tableCount)
	}
}

func testMasterSlaveReplication() {
	// 连接主库和从库
	masterDB, err := sql.Open("mysql", "chatuser:chatpassword@tcp(localhost:3316)/chat_app")
	if err != nil {
		fmt.Printf("   ❌ 连接主库失败: %v\n", err)
		return
	}
	defer masterDB.Close()

	slave1DB, err := sql.Open("mysql", "chatuser:chatpassword@tcp(localhost:3317)/chat_app")
	if err != nil {
		fmt.Printf("   ❌ 连接从库1失败: %v\n", err)
		return
	}
	defer slave1DB.Close()

	// 在主库插入测试数据
	testUsername := fmt.Sprintf("system_test_%d", time.Now().Unix())
	testEmail := fmt.Sprintf("%s@test.com", testUsername)

	fmt.Printf("   📝 在主库插入测试用户: %s\n", testUsername)

	_, err = masterDB.Exec(`
		INSERT INTO users (username, email, password, password_hash, avatar_url, status, is_online) 
		VALUES (?, ?, 'system_test_password', 'system_test_hash', NULL, 'offline', 0)
	`, testUsername, testEmail)

	if err != nil {
		fmt.Printf("   ❌ 插入数据失败: %v\n", err)
		return
	}

	// 等待复制同步
	fmt.Println("   ⏳ 等待主从复制同步...")
	time.Sleep(2 * time.Second)

	// 检查从库数据
	var count int
	err = slave1DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", testUsername).Scan(&count)
	if err != nil {
		fmt.Printf("   ❌ 从库查询失败: %v\n", err)
		return
	}

	if count > 0 {
		fmt.Printf("   ✅ 主从复制正常: 从库1已同步数据\n")
	} else {
		fmt.Printf("   ❌ 主从复制异常: 从库1未同步数据\n")
	}

	// 查询总用户数对比
	var masterCount, slaveCount int
	masterDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&masterCount)
	slave1DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&slaveCount)

	fmt.Printf("   📊 用户数对比 - 主库: %d, 从库1: %d\n", masterCount, slaveCount)
}

func testRedisConnection() {
	conn, err := net.DialTimeout("tcp", "localhost:6379", 3*time.Second)
	if err != nil {
		fmt.Printf("   ❌ Redis连接失败: %v\n", err)
		return
	}
	defer conn.Close()

	// 发送简单的PING命令
	_, err = conn.Write([]byte("PING\r\n"))
	if err != nil {
		fmt.Printf("   ❌ Redis PING失败: %v\n", err)
		return
	}

	// 读取响应
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("   ❌ Redis响应读取失败: %v\n", err)
		return
	}

	response := string(buffer[:n])
	if strings.Contains(response, "PONG") {
		fmt.Printf("   ✅ Redis连接成功: %s", strings.TrimSpace(response))
	} else {
		fmt.Printf("   ❌ Redis响应异常: %s", strings.TrimSpace(response))
	}
}
