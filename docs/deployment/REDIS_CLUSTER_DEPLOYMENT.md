# Redis Cluster 部署指南

## 概述

本文档描述如何在聊天系统中部署和使用 Redis Cluster。Redis Cluster 提供了高可用性、自动分片和故障转移功能。

## 系统架构

```
聊天系统架构:
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Chat Client   │    │   Chat Client   │    │   Chat Client   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                   ┌─────────────────┐
                   │  Chat Server    │
                   │  (Zinx TCP)     │
                   └─────────────────┘
                            │
                ┌───────────┼───────────┐
                │                       │
    ┌─────────────────┐    ┌─────────────────┐
    │  MySQL Cluster  │    │  Redis Cluster  │
    │  (主从 + 分片)    │    │  (3主3从)       │
    └─────────────────┘    └─────────────────┘
```

## Redis Cluster 配置

### 集群节点配置

- **Master 节点**: 3个 (redis-master-1:7001, redis-master-2:7002, redis-master-3:7003)
- **Slave 节点**: 3个 (redis-slave-1:7004, redis-slave-2:7005, redis-slave-3:7006)
- **复制因子**: 1 (每个主节点有1个从节点)
- **哈希槽**: 16384个槽位自动分配

### 配置文件说明

#### 1. 服务器配置 (`chat-server/conf/config.json`)
```json
{
  "Database": {
    "Redis": {
      "ClusterEnabled": true,
      "ClusterAddrs": [
        "redis-master-1:7001",
        "redis-master-2:7002", 
        "redis-master-3:7003",
        "redis-slave-1:7004",
        "redis-slave-2:7005",
        "redis-slave-3:7006"
      ],
      "PoolSize": 50,
      "MinIdleConns": 20,
      "MaxRetries": 3,
      "MessageExpiration": 604800
    }
  }
}
```

#### 2. Docker Compose 配置 (`docker-compose.yml`)
包含完整的 Redis Cluster 服务定义，包括：
- 6个 Redis 节点容器
- 自动集群初始化容器
- 数据持久化卷
- 网络配置

## 部署步骤

### Windows PowerShell 快速部署

```powershell
# 1. 确保 Docker Desktop 已安装并运行
docker --version
docker-compose --version

# 2. 启动整个系统
.\start-cluster.ps1

# 3. 检查状态
.\check-cluster.ps1
```

### Linux/macOS 快速部署

```bash
# 1. 确保 Docker 和 Docker Compose 已安装
docker --version
docker-compose --version

# 2. 给启动脚本执行权限
chmod +x start-cluster.sh
chmod +x check-cluster.sh

# 3. 启动整个系统
./start-cluster.sh
```

### 手动部署 (通用)

```bash
# 1. 创建网络
docker network create chat-network

# 2. 生成 Redis 配置文件
cd redis-cluster
chmod +x create-configs.sh
./create-configs.sh
cd ..

# 3. 启动 MySQL 服务
docker-compose up -d mysql-master mysql-slave1 mysql-slave2 mysql-shard0 mysql-shard1

# 4. 启动 Redis 节点
docker-compose up -d redis-master-1 redis-master-2 redis-master-3
docker-compose up -d redis-slave-1 redis-slave-2 redis-slave-3

# 5. 等待节点启动
sleep 15

# 6. 初始化集群
docker-compose up -d redis-cluster-init

# 7. 启动聊天服务器
docker-compose up -d chat-server

# 8. 启动监控服务
docker-compose up -d adminer prometheus grafana
```

## 验证部署

### 1. 检查集群状态

**Windows PowerShell:**
```powershell
# 使用检查脚本
.\check-cluster.ps1

# 或手动检查
docker exec redis-master-1 redis-cli -p 7001 cluster info
docker exec redis-master-1 redis-cli -p 7001 cluster nodes
```

**Linux/macOS:**
```bash
# 使用检查脚本
./check-cluster.sh

# 或手动检查
docker exec redis-master-1 redis-cli -p 7001 cluster info
docker exec redis-master-1 redis-cli -p 7001 cluster nodes
```

### 2. 测试集群功能

```bash
# 测试写入
docker exec redis-master-1 redis-cli -c -p 7001 set test-key "Hello Cluster"

# 测试读取
docker exec redis-master-1 redis-cli -c -p 7001 get test-key

# 测试不同节点
docker exec redis-master-2 redis-cli -c -p 7002 get test-key
```

### 3. 查看聊天服务器日志

```bash
docker-compose logs chat-server
```

应该看到类似输出：
```
初始化Redis连接...
Redis集群模式连接成功!
集群节点: [redis-master-1:7001 redis-master-2:7002 redis-master-3:7003 redis-slave-1:7004 redis-slave-2:7005 redis-slave-3:7006]
所有服务初始化完毕!
```

## 功能特性

### 1. 自动故障转移
- 当主节点失效时，对应的从节点会自动升级为主节点
- 集群继续提供服务，无需人工干预

### 2. 数据分片
- 数据自动分布在3个主节点上
- 使用一致性哈希算法分配数据

### 3. 高可用性
- 任意单点故障不会影响整体服务
- 支持动态添加/删除节点

### 4. 性能优化
- 连接池管理 (默认50个连接)
- 自动重试机制 (默认3次)
- 读写分离支持

## 监控和管理

### 1. 访问地址

- **聊天服务器**: localhost:9000 (TCP), localhost:8080 (HTTP)
- **数据库管理**: http://localhost:8081
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)

### 2. Redis 节点访问

- **Master 1**: localhost:7001
- **Master 2**: localhost:7002
- **Master 3**: localhost:7003
- **Slave 1**: localhost:7004
- **Slave 2**: localhost:7005
- **Slave 3**: localhost:7006

### 3. 集群管理命令

```bash
# 查看集群信息
docker exec redis-master-1 redis-cli -p 7001 cluster info

# 查看节点状态
docker exec redis-master-1 redis-cli -p 7001 cluster nodes

# 查看插槽分布
docker exec redis-master-1 redis-cli -p 7001 cluster slots

# 添加节点 (示例)
docker exec redis-master-1 redis-cli -p 7001 cluster meet <new-node-ip> <new-node-port>

# 重新分片 (示例)
docker exec redis-master-1 redis-cli --cluster reshard redis-master-1:7001
```

## 故障排除

### 1. 集群初始化失败

```bash
# 检查所有节点是否运行
docker ps | grep redis

# 检查网络连接
docker exec redis-master-1 ping redis-master-2

# 手动初始化集群
docker exec redis-master-1 redis-cli --cluster create \
  redis-master-1:7001 redis-master-2:7002 redis-master-3:7003 \
  redis-slave-1:7004 redis-slave-2:7005 redis-slave-3:7006 \
  --cluster-replicas 1 --cluster-yes
```

### 2. 聊天服务器连接失败

```bash
# 检查配置文件
cat chat-server/conf/config.json

# 检查服务器日志
docker-compose logs chat-server

# 重启服务器
docker-compose restart chat-server
```

### 3. 性能问题

```bash
# 检查连接池状态
docker exec redis-master-1 redis-cli -p 7001 info clients

# 检查内存使用
docker exec redis-master-1 redis-cli -p 7001 info memory

# 检查网络延迟
docker exec chat-server ping redis-master-1
```

## 生产环境建议

### 1. 安全配置
- 设置 Redis 密码
- 配置防火墙规则
- 使用 TLS/SSL 加密

### 2. 性能调优
- 根据负载调整连接池大小
- 配置适当的超时时间
- 监控内存使用情况

### 3. 备份策略
- 定期备份 RDB 文件
- 配置 AOF 持久化
- 实施跨地域复制

### 4. 监控告警
- 设置节点健康检查
- 配置内存使用告警
- 监控连接数和延迟

## 常用命令

**Windows PowerShell:**
```powershell
# 启动系统
.\start-cluster.ps1

# 检查状态
.\check-cluster.ps1

# 查看日志
docker-compose logs -f chat-server

# 停止系统
docker-compose down

# 清理数据 (谨慎使用)
docker-compose down -v
```

**Linux/macOS:**
```bash
# 启动系统
./start-cluster.sh

# 检查状态
./check-cluster.sh

# 查看日志
docker-compose logs -f chat-server

# 停止系统
docker-compose down

# 清理数据 (谨慎使用)
docker-compose down -v
``` 