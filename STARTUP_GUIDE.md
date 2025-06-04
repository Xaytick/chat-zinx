# 🚀 Chat-Zinx 分布式聊天系统启动指南

## 📋 环境要求

- **Windows 10/11**
- **Docker Desktop** (已验证 v28.0.4)
- **Docker Compose** (已验证 v2.34.0)
- **PowerShell 5.0+**
- **Go 1.22+** (可选，用于编译)

## 🎯 快速启动（推荐）

### 第一步：启动分布式系统

```powershell
# 进入项目目录
cd D:\code\chat-zinx

# 方式1：一键启动完整分布式系统（推荐）
.\scripts\deployment\start-distributed-system.ps1

# 方式2：清理并重新启动
.\scripts\deployment\start-distributed-system.ps1 -Clean

# 方式3：启动并查看实时日志
.\scripts\deployment\start-distributed-system.ps1 -Logs

# 方式4：自定义服务器数量
.\scripts\deployment\start-distributed-system.ps1 -ServerCount 5
```

这个脚本会自动启动：
- ✅ **NATS集群** (3节点) - 消息传递
- ✅ **Consul集群** (3节点) - 服务发现
- ✅ **Redis集群** (6节点) - 数据存储
- ✅ **MySQL主从** (5节点) - 数据库
- ✅ **多聊天服务器** (默认3个) - 应用服务
- ✅ **监控系统** - Prometheus + Grafana

### 第二步：验证系统状态

```powershell
# 检查所有服务状态
.\scripts\deployment\check-cluster.ps1

# 手动检查各服务
docker ps
```

## 🌐 系统访问地址

### Web界面
- **Consul管理界面**: http://localhost:8500
- **NATS监控界面**: http://localhost:8222
- **Grafana监控**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090

### 聊天服务器
- **Chat-Server-1**: TCP 9000, HTTP 8080
- **Chat-Server-2**: TCP 9001, HTTP 8081  
- **Chat-Server-3**: TCP 9002, HTTP 8082

### 数据库
- **MySQL主库**: localhost:3316
- **Redis集群**: localhost:7001-7006

## 💬 测试聊天功能

### 方法一：使用专用测试客户端

```powershell
# 编译测试客户端
cd scripts\tools
go build -o test_client.exe test_distributed_messaging.go

# 启动客户端A（连接Server1）
.\test_client.exe
# 选择服务器: 1 (localhost:9000)
# 用户名: alice
# 密码: password123

# 启动客户端B（连接Server2）
.\test_client.exe  
# 选择服务器: 2 (localhost:9001)
# 用户名: bob
# 密码: password123

# 在客户端A中发送消息给B
[alice@localhost:9000] > p bob Hello from Server1!

# 在客户端B中应该收到
📨 [P2P消息] 来自 alice: Hello from Server1!
```

### 方法二：使用Telnet测试

```powershell
# 终端1：连接Server1
telnet localhost 9000

# 终端2：连接Server2
telnet localhost 9001

# 分别登录不同用户，测试跨服务器通信
```

## 🧪 测试场景

### 1. 跨服务器P2P消息
```powershell
# 用户A连接Server1，用户B连接Server2
# 测试A向B发送消息，验证NATS路由是否正常
```

### 2. 分布式群组消息
```powershell
# 多个用户分布在不同服务器
# 测试群组消息广播功能
[alice@Server1] > g 1 大家好！
```

### 3. 服务器故障转移
```powershell
# 停止一个服务器，验证用户是否能连接到其他服务器
docker stop chat-server-1
```

## 📊 系统监控

### Consul监控（推荐）
1. 打开 http://localhost:8500
2. 查看 **Services** 页面，确认所有chat-server已注册
3. 查看 **Key/Value** 页面，监控用户在线状态

### NATS监控
1. 打开 http://localhost:8222
2. 查看连接数和消息统计
3. 监控Subject订阅情况

### 命令行监控
```powershell
# 查看用户分布
curl http://localhost:8500/v1/kv/users/online/?recurse

# 查看NATS统计
curl http://localhost:8222/varz

# 查看服务健康状态  
curl http://localhost:8080/health
```

## 🛠️ 常见操作

### 重启单个服务
```powershell
# 重启聊天服务器
docker restart chat-server-1

# 重启Redis节点
docker restart redis-1

# 重启NATS节点
docker restart nats-1
```

### 查看日志
```powershell
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker logs chat-server-1 -f
docker logs nats-1 -f
docker logs consul-1 -f
```

### 扩展服务器
```powershell
# 启动更多聊天服务器实例
.\scripts\deployment\start-distributed-system.ps1 -ServerCount 5
```

## 🔧 故障排查

### 1. 容器启动失败
```powershell
# 检查Docker资源
docker system df
docker system prune -f

# 检查端口占用
netstat -an | findstr :9000
```

### 2. NATS连接失败
```powershell
# 检查NATS集群状态
docker logs nats-1
telnet localhost 4222
```

### 3. 服务注册失败
```powershell
# 检查Consul状态
curl http://localhost:8500/v1/status/leader
docker logs consul-1
```

### 4. 跨服务器消息不通
```powershell
# 检查用户在线状态
curl http://localhost:8500/v1/kv/users/online/?recurse

# 检查NATS订阅
curl http://localhost:8222/subsz
```

## 🛑 停止系统

```powershell
# 停止所有服务
docker-compose down

# 停止并删除所有数据
docker-compose down -v

# 清理分布式组件
.\scripts\deployment\start-distributed-system.ps1 -Clean
```

## 📚 更多文档

- [分布式消息传递解决方案](docs/architecture/DISTRIBUTED_MESSAGING_SOLUTION.md)
- [分布式消息系统指南](docs/guides/DISTRIBUTED_MESSAGING_GUIDE.md)
- [Redis集群指南](docs/guides/REDIS_CLUSTER_GUIDE.md)

## 🎉 成功标志

当您看到以下输出时，系统启动成功：

```
=== 分布式聊天系统启动完成! ===
系统现在支持:
  ✓ 跨服务器P2P消息传递
  ✓ 分布式群组消息广播  
  ✓ 服务自动发现与健康检查
  ✓ 高可用性和故障转移
  ✓ 水平扩展支持
```

享受您的分布式聊天系统！🚀 