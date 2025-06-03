# Chat-Zinx 聊天系统

基于Zinx框架构建的高性能分布式聊天系统，支持MySQL集群和Redis集群。

## 🚀 项目特性

- **高性能**: 基于Zinx网络框架，支持高并发连接
- **分布式**: MySQL主从复制 + 分片，Redis集群架构
- **监控**: 集成Prometheus + Grafana监控系统
- **容器化**: 完全Docker化部署，一键启动
- **可扩展**: 模块化设计，易于扩展功能

## 📁 项目结构

```
chat-zinx/
├── README.md                           # 项目说明
├── docker-compose.yml                  # Docker编排文件
├── go.mod / go.sum                     # Go模块管理
│
├── docs/                              # 📚 文档
│   ├── deployment/                    # 部署文档
│   ├── guides/                        # 操作指南
│   └── architecture/                  # 架构文档
│
├── scripts/                           # 🛠️ 脚本工具
│   ├── deployment/                    # 部署脚本
│   ├── database/                      # 数据库脚本
│   └── tools/                         # 工具脚本
│
├── infrastructure/                    # 🏗️ 基础设施
│   ├── redis-cluster/                 # Redis集群配置
│   ├── monitoring/                    # 监控配置
│   └── mysql/                         # MySQL配置
│
├── chat-server/                       # 💬 聊天服务器
├── chat-client/                       # 👤 聊天客户端
└── tests/                             # 🧪 测试文件
```

## ⚡ 快速开始

### 1. 启动集群

```powershell
# Windows PowerShell
.\scripts\deployment\start-cluster.ps1

# 初始化Redis集群（如需要）
.\scripts\deployment\init-cluster.ps1
```

### 2. 检查状态

```powershell
.\scripts\deployment\check-cluster.ps1
```

### 3. 访问服务

- **聊天服务器**: TCP端口 9000, HTTP端口 8080
- **MySQL主库**: localhost:3316
- **Redis集群**: localhost:7001-7006
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)

## 🔧 配置说明

### MySQL集群
- **主库**: chat-mysql-master:3306
- **从库**: chat-mysql-slave1:3306, chat-mysql-slave2:3306
- **分片**: chat-mysql-shard0:3306, chat-mysql-shard1:3306

### Redis集群
- **主节点**: redis-master-1:7001, redis-master-2:7002, redis-master-3:7003
- **从节点**: redis-slave-1:7004, redis-slave-2:7005, redis-slave-3:7006
- **架构**: 3主3从，自动分片和故障转移

## 📖 文档导航

- [Redis集群部署指南](docs/deployment/REDIS_CLUSTER_DEPLOYMENT.md)
- [系统部署成功记录](docs/deployment/DEPLOYMENT_SUCCESS.md)
- [Redis集群操作指南](docs/guides/REDIS_CLUSTER_GUIDE.md)
- [项目重构计划](docs/PROJECT_RESTRUCTURE_PLAN.md)

## 🛠️ 开发工具

```powershell
# 系统检查
go run scripts/tools/system_check.go

# Redis集群测试
go run scripts/tools/test_cluster.go

# 查看容器状态
docker ps

# 查看服务日志
docker-compose logs [service-name]
```

## 📊 监控

系统集成了完整的监控解决方案：

- **Prometheus**: 指标收集和存储
- **Grafana**: 可视化仪表板
- **监控指标**: CPU、内存、网络、Redis、MySQL等

## 🤝 贡献

欢迎提交Issue和Pull Request来改进项目！

## �� 许可证

MIT License 