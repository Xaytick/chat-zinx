# Redis Cluster 一键部署脚本 (PowerShell)

Write-Host "🚀 开始部署Redis Cluster..." -ForegroundColor Green

# 检查Docker网络
$networkExists = docker network ls | Select-String "chat-network"
if (-not $networkExists) {
    Write-Host "📡 创建Docker网络..." -ForegroundColor Yellow
    docker network create chat-network
}

# 停止并清理现有容器
Write-Host "🧹 清理现有Redis容器..." -ForegroundColor Yellow
docker stop redis-master-1 redis-master-2 redis-master-3 redis-slave-1 redis-slave-2 redis-slave-3 redis-cluster-init 2>$null
docker rm redis-master-1 redis-master-2 redis-master-3 redis-slave-1 redis-slave-2 redis-slave-3 redis-cluster-init 2>$null

# 启动Redis Cluster
Write-Host "🔧 启动Redis Cluster..." -ForegroundColor Yellow
docker-compose -f docker-compose-redis.yml up -d

# 等待容器启动
Write-Host "⏳ 等待Redis节点启动..." -ForegroundColor Yellow
Start-Sleep -Seconds 15

# 检查节点状态
Write-Host "🔍 检查节点状态..." -ForegroundColor Yellow
$ports = @(7001, 7002, 7003, 7004, 7005, 7006)
foreach ($port in $ports) {
    try {
        $result = docker exec redis-master-1 redis-cli -h localhost -p $port ping 2>$null
        if ($result -eq "PONG") {
            Write-Host "✅ 节点 $port 运行正常" -ForegroundColor Green
        } else {
            Write-Host "❌ 节点 $port 连接失败" -ForegroundColor Red
        }
    } catch {
        Write-Host "❌ 节点 $port 连接失败" -ForegroundColor Red
    }
}

# 创建集群
Write-Host "🔗 创建Redis Cluster..." -ForegroundColor Yellow
docker exec redis-master-1 redis-cli --cluster create redis-master-1:7001 redis-master-2:7002 redis-master-3:7003 redis-slave-1:7004 redis-slave-2:7005 redis-slave-3:7006 --cluster-replicas 1 --cluster-yes

# 验证集群状态
Write-Host "📊 验证集群状态..." -ForegroundColor Yellow
docker exec redis-master-1 redis-cli -p 7001 cluster info
docker exec redis-master-1 redis-cli -p 7001 cluster nodes

Write-Host "🎉 Redis Cluster 部署完成！" -ForegroundColor Green
Write-Host ""
Write-Host "📋 集群信息:" -ForegroundColor Cyan
Write-Host "   主节点: redis-master-1:7001, redis-master-2:7002, redis-master-3:7003"
Write-Host "   从节点: redis-slave-1:7004, redis-slave-2:7005, redis-slave-3:7006"
Write-Host "   数据分片: 16384个槽位平均分配到3个主节点"
Write-Host "   高可用: 每个主节点有1个从节点备份"
Write-Host ""
Write-Host "🔧 管理命令:" -ForegroundColor Cyan
Write-Host "   查看集群状态: docker exec redis-master-1 redis-cli -p 7001 cluster info"
Write-Host "   查看节点信息: docker exec redis-master-1 redis-cli -p 7001 cluster nodes"
Write-Host "   连接集群: docker exec -it redis-master-1 redis-cli -c -p 7001"
Write-Host ""
Write-Host "🧪 测试集群:" -ForegroundColor Cyan
Write-Host "   写入数据: docker exec redis-master-1 redis-cli -c -p 7001 set test_key 'Hello'"
Write-Host "   读取数据: docker exec redis-master-1 redis-cli -c -p 7001 get test_key" 