# ChatZinx 客户端

本客户端是 ChatZinx 项目的组成部分，旨在提供一个与 ChatZinx 服务器进行交互的命令行或嵌入式客户端。

## 功能特性

* 用户注册与登录
* 发送和接收私聊消息
* （计划中）群组聊天功能：创建群组、加入群组、离开群组、发送群聊消息
* （计划中）查看历史消息
* 通过自定义协议与服务器高效通信

## 项目结构

```
chat-client/
├── pkg/
│   ├── client/          # 客户端核心逻辑 (连接管理, 消息收发, 用户认证等)
│   └── protocol/      # 与服务器共享的通信协议定义 (消息ID, 数据结构等)
└── main.go              # (示例) 客户端主程序入口 (如果提供一个独立的客户端应用)
```

*注意：具体的 `main.go` 或示例客户端的实现可能因开发阶段而异。`pkg/client` 提供了核心的客户端库功能。*

## 先决条件

* Go 1.20 或更高版本
* ChatZinx 服务器正在运行并可访问

## 配置

客户端通常需要配置以下信息：

* **服务器地址**: ChatZinx 服务器的 IP 地址和端口 (例如: `127.0.0.1:9000`)

具体的配置方式可能通过命令行参数、配置文件或代码内硬编码实现，取决于客户端的具体实现。

## 如何运行 (示例)

由于客户端的具体实现可能多样 (例如，作为一个库被其他应用导入，或作为一个独立的命令行工具)，以下是一个通用指南：

1. **获取代码**:
   如果您是项目的开发者，代码已在您的工作区中。
2. **编译和运行**:
   如果提供了一个示例 `main.go`：

   ```bash
   cd chat-client
   go build -o chat_client main.go
   ./chat_client --server_addr="127.0.0.1:9000"
   ```

   或者直接运行：

   ```bash
   cd chat-client
   go run main.go --server_addr="127.0.0.1:9000"
   ```

   如果 `pkg/client` 主要作为库使用，您需要在您的 Go 项目中导入它：

   ```go
   import "github.com/Xaytick/chat-zinx/chat-client/pkg/client"
   // ... 然后使用 client 包的功能
   ```

## 核心组件

* `pkg/client`: 封装了与服务器交互的底层逻辑。
  * 连接建立与维护
  * 用户认证流程 (注册、登录)
  * 消息的序列化、发送、接收和反序列化
  * 提供 API 供上层应用调用。
* `pkg/protocol`: 定义了客户端和服务器之间的消息格式和类型。这对于确保双方正确通信至关重要。

## 如何使用 (作为库)

以下是如何在您的代码中使用 `pkg/client` 包的伪代码示例：

```go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Xaytick/chat-zinx/chat-client/pkg/client" // 假设的导入路径
	"github.com/Xaytick/chat-zinx/chat-client/pkg/protocol" // 假设的导入路径
)

func main() {
	// 1. 配置服务器地址
	serverAddr := "127.0.0.1:9000" // 从配置或命令行获取

	// 2. 创建客户端实例
	chatCli, err := client.NewClient(serverAddr) // NewClient 是一个假设的构造函数
	if err != nil {
		fmt.Println("无法连接到服务器:", err)
		return
	}
	defer chatCli.Close()

	// 3. 启动消息监听器 (处理来自服务器的消息)
	go chatCli.StartReceiving(func(msgID uint32, data []byte) {
		fmt.Printf("\n[服务器消息 ID: %d]: %s\n", msgID, string(data)) // 示例处理
        // 根据 msgID 解析 data 为具体的响应结构体
        // 例如：
        // if msgID == protocol.MsgIdLoginRes {
        //     var res protocol.LoginRes
        //     // 解析 data 到 res
        //     fmt.Println("登录结果:", res.Error)
        // }
	})

	// 4. 用户交互 (示例：简单的命令行交互)
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("客户端已启动。输入 'register <用户名> <密码>' 或 'login <用户名> <密码>'")

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		parts := strings.Fields(input)

		if len(parts) == 0 {
			continue
		}

		command := parts[0]

		switch command {
		case "register":
			if len(parts) == 3 {
				// err := chatCli.Register(parts[1], parts[2]) // Register 是一个假设的方法
				// if err != nil {
				// 	fmt.Println("注册失败:", err)
				// } else {
				// 	fmt.Println("注册请求已发送")
				// }
				fmt.Println("注册功能 (示例，具体实现依赖 client 包)")
			} else {
				fmt.Println("用法: register <用户名> <密码>")
			}
		case "login":
			if len(parts) == 3 {
				// err := chatCli.Login(parts[1], parts[2]) // Login 是一个假设的方法
				// if err != nil {
				//  	fmt.Println("登录失败:", err)
				// } else {
				//  	fmt.Println("登录请求已发送")
				// }
				fmt.Println("登录功能 (示例，具体实现依赖 client 包)")
			} else {
				fmt.Println("用法: login <用户名> <密码>")
			}
		case "send": // send <接收者ID> <消息内容>
			// if len(parts) >= 3 {
			//  	targetUserID := parts[1]
			//  	message := strings.Join(parts[2:], " ")
			//  	// err := chatCli.SendPrivateMessage(targetUserID, message) // SendPrivateMessage 是一个假设的方法
			//  	// if err != nil {
			//  	//	fmt.Println("发送失败:", err)
			//  	// }
			//		fmt.Println("发送私聊消息 (示例)")
			// } else {
			//  	fmt.Println("用法: send <接收者ID> <消息内容>")
			// }
			fmt.Println("发送消息功能 (示例，具体实现依赖 client 包)")
		case "exit":
			fmt.Println("正在关闭客户端...")
			return
		default:
			fmt.Println("未知命令。可用命令: register, login, send, exit")
		}
	}
}
```

## 贡献

如果您想为该项目做出贡献，请遵循标准的 GitHub Fork 和 Pull Request 工作流程。

---

请注意：上述 `main.go` 示例和 `pkg/client` 的使用方法是基于对典型客户端库功能的推测。需要根据 `pkg/client` 的实际 API 进行调整。
