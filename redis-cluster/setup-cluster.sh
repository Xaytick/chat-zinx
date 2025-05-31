#!/bin/bash

echo "🚀 开始部署Redis Cluster..."

# 检查Docker网络
if ! docker network ls | grep -q "chat-network"; then
    echo "📡 创建Docker网络..."
    docker network create chat-network
fi

# 停止并清理现有的Redis容器
echo "🧹 清理现有Redis容器..."
docker stop redis-master-1 redis-master-2 redis-master-3 \
           redis-slave-1 redis-slave-2 redis-slave-3 \
           redis-cluster-init 2>/dev/null || true

docker rm redis-master-1 redis-master-2 redis-master-3 \
         redis-slave-1 redis-slave-2 redis-slave-3 \
         redis-cluster-init 2>/dev/null || true

# 启动Redis Cluster
echo "🔧 启动Redis Cluster..."
docker-compose -f docker-compose-redis.yml up -d

# 等待容器启动
echo "⏳ 等待Redis节点启动..."
sleep 15

# 检查节点状态
echo "🔍 检查节点状态..."
for port in 7001 7002 7003 7004 7005 7006; do
    if docker exec redis-master-1 redis-cli -h localhost -p $port ping 2>/dev/null | grep -q PONG; then
        echo "✅ 节点 $port 运行正常"
    else
        echo "❌ 节点 $port 连接失败"
    fi
done

# 创建集群
echo "🔗 创建Redis Cluster..."
docker exec redis-master-1 redis-cli --cluster create \
    redis-master-1:7001 redis-master-2:7002 redis-master-3:7003 \
    redis-slave-1:7004 redis-slave-2:7005 redis-slave-3:7006 \
    --cluster-replicas 1 --cluster-yes

# 验证集群状态
echo "📊 验证集群状态..."
docker exec redis-master-1 redis-cli -c -h redis-master-1 -p 7001 cluster info
docker exec redis-master-1 redis-cli -c -h redis-master-1 -p 7001 cluster nodes

echo "🎉 Redis Cluster 部署完成！"
echo ""
echo "📋 集群信息:"
echo "   主节点: redis-master-1:7001, redis-master-2:7002, redis-master-3:7003"
echo "   从节点: redis-slave-1:7004, redis-slave-2:7005, redis-slave-3:7006"
echo "   数据分片: 16384个槽位平均分配到3个主节点"
echo "   高可用: 每个主节点有1个从节点备份"
echo ""
echo "🔧 管理命令:"
echo "   查看集群状态: docker exec redis-master-1 redis-cli -c cluster info"
echo "   查看节点信息: docker exec redis-master-1 redis-cli -c cluster nodes"
echo "   连接集群: docker exec -it redis-master-1 redis-cli -c" 