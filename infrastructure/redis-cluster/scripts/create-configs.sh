#!/bin/bash

# 创建Redis Cluster配置文件脚本

# 定义节点信息
declare -A nodes=(
    ["7001"]="redis-master-1"
    ["7002"]="redis-master-2"
    ["7003"]="redis-master-3"
    ["7004"]="redis-slave-1"
    ["7005"]="redis-slave-2"
    ["7006"]="redis-slave-3"
)

# 为每个节点创建配置文件
for port in "${!nodes[@]}"; do
    node_name="${nodes[$port]}"
    bus_port=$((port + 10000))
    
    cat > redis-configs/redis-${port}.conf << EOF
# Redis Cluster 节点 ${node_name} 配置
port ${port}
bind 0.0.0.0

# 集群配置
cluster-enabled yes
cluster-config-file nodes-${port}.conf
cluster-node-timeout 15000
cluster-announce-ip ${node_name}
cluster-announce-port ${port}
cluster-announce-bus-port ${bus_port}

# 持久化配置
dir /data
appendonly yes
appendfsync everysec
save 900 1
save 300 10
save 60 10000

# 内存配置
maxmemory 1gb
maxmemory-policy allkeys-lru

# 网络配置
timeout 0
tcp-keepalive 300

# 日志配置
loglevel notice
logfile ""

# 安全配置
protected-mode no

# 性能优化
tcp-backlog 511
databases 1  # 集群模式只支持单个数据库
EOF

    echo "创建配置文件: redis-${port}.conf (${node_name})"
done

echo "所有Redis Cluster配置文件创建完成！" 