# 聊天客户端

## 项目结构

```
chat-client/
├── client1/             # 客户端1
│   └── main.go          # 客户端1入口
├── client2/             # 客户端2
│   └── main.go          # 客户端2入口
└── pkg/                 # 共享包
    └── client/          # 客户端共享功能
        └── client.go    # 客户端核心实现
```

## 共享客户端包

为了避免代码重复，我们创建了一个共享的客户端包 `pkg/client`，它提供了以下功能：

- 连接服务器
- 用户注册与登录
- 发送和接收消息
- 消息处理

## 使用方法

```go
// 创建客户端实例
cli, err := client.NewChatClient("127.0.0.1:9000")
if err != nil {
    panic(err)
}
defer cli.Close()

// 注册并登录用户
if err := cli.RegisterAndLogin("username", "password"); err != nil {
    panic(err)
}

// 发送消息
if err := cli.SendTextMessage("targetUser", "Hello!"); err != nil {
    panic(err)
}

// 处理接收到的消息
cli.StartMsgListener(func(msgID uint32, msgBody []byte) {
    // 处理消息...
})
```

## 客户端功能

- **client1**: 以 testuser1 身份登录，发送消息给 testuser2
- **client2**: 以 testuser2 身份登录，发送消息给 testuser1

## 运行方法

1. 确保服务器已启动
2. 启动客户端1: `cd client1 && go run main.go`
3. 启动客户端2: `cd client2 && go run main.go`

两个客户端将自动注册（如果用户不存在）、登录，并可以互相发送消息。 