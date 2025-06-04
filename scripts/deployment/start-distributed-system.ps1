# ============================================================================
# Chat-Zinx 分布式聊天系统启动脚本
# 此脚本将启动完整的分布式聊天系统，包括：
# - NATS集群
# - Consul集群  
# - Redis集群
# - MySQL集群
# - 多个聊天服务器实例
# - 监控系统
# ============================================================================

param(
    [switch]$Clean,
    [switch]$Logs,
    [int]$ServerCount = 3
)

Write-Host "=== Chat-Zinx 分布式聊天系统启动脚本 ===" -ForegroundColor Green

# 设置错误处理
$ErrorActionPreference = "Stop"

# 获取项目根目录
$ProjectRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
Set-Location $ProjectRoot

Write-Host "项目根目录: $ProjectRoot" -ForegroundColor Yellow

# 清理旧容器（如果指定）
if ($Clean) {
    Write-Host "清理现有容器..." -ForegroundColor Yellow
    
    # 停止并删除所有相关容器
    $containers = @(
        "nats-1", "nats-2", "nats-3",
        "consul-1", "consul-2", "consul-3", 
        "nginx-lb",
        "chat-server-1", "chat-server-2", "chat-server-3"
    )
    
    foreach ($container in $containers) {
        try {
            docker stop $container 2>$null
            docker rm $container 2>$null
            Write-Host "删除容器: $container" -ForegroundColor Gray
        } catch {
            # 忽略错误，容器可能不存在
        }
    }
    
    # 清理网络
    try {
        docker network rm chat-network 2>$null
        Write-Host "删除网络: chat-network" -ForegroundColor Gray
    } catch {
        # 忽略错误
    }
}

# 创建Docker网络
Write-Host "创建Docker网络..." -ForegroundColor Yellow
try {
    docker network create chat-network 2>$null
    Write-Host "✓ 创建网络 chat-network" -ForegroundColor Green
} catch {
    Write-Host "网络 chat-network 已存在" -ForegroundColor Gray
}

# 第一步：启动基础设施 (NATS + Consul)
Write-Host "`n=== 第一步：启动基础设施 ===" -ForegroundColor Green

Write-Host "启动 NATS + Consul 集群..." -ForegroundColor Yellow
docker-compose -f infrastructure/docker-compose-messaging.yml up -d

# 等待基础设施启动
Write-Host "等待基础设施启动完成..." -ForegroundColor Yellow
Start-Sleep 15

# 检查基础设施状态
Write-Host "检查基础设施状态..." -ForegroundColor Yellow

# 检查NATS集群
$natsHealthy = $true
for ($i = 1; $i -le 3; $i++) {
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:$((8221 + $i))/healthz" -TimeoutSec 5
        Write-Host "✓ NATS-$i 运行正常" -ForegroundColor Green
    } catch {
        Write-Host "✗ NATS-$i 状态异常" -ForegroundColor Red
        $natsHealthy = $false
    }
}

# 检查Consul集群
$consulHealthy = $true
for ($i = 1; $i -le 3; $i++) {
    try {
        $port = if ($i -eq 1) { 8500 } else { 8500 + $i }
        $response = Invoke-RestMethod -Uri "http://localhost:$port/v1/status/leader" -TimeoutSec 5
        Write-Host "✓ Consul-$i 运行正常" -ForegroundColor Green
    } catch {
        Write-Host "✗ Consul-$i 状态异常" -ForegroundColor Red
        $consulHealthy = $false
    }
}

if (-not $natsHealthy -or -not $consulHealthy) {
    Write-Host "基础设施启动失败，请检查日志!" -ForegroundColor Red
    exit 1
}

# 第二步：启动现有服务 (Redis + MySQL)
Write-Host "`n=== 第二步：启动现有服务 ===" -ForegroundColor Green

Write-Host "启动 Redis 集群..." -ForegroundColor Yellow
docker-compose up -d

# 等待Redis集群启动
Write-Host "等待 Redis 集群启动..." -ForegroundColor Yellow
Start-Sleep 10

# 检查Redis集群状态
Write-Host "检查 Redis 集群状态..." -ForegroundColor Yellow
try {
    $redisInfo = docker exec redis-1 redis-cli cluster info
    if ($redisInfo -match "cluster_state:ok") {
        Write-Host "✓ Redis 集群运行正常" -ForegroundColor Green
    } else {
        Write-Host "✗ Redis 集群状态异常" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ 无法连接到 Redis 集群" -ForegroundColor Red
}

# 第三步：构建聊天服务器镜像
Write-Host "`n=== 第三步：构建聊天服务器镜像 ===" -ForegroundColor Green

Set-Location "chat-server"

# 检查是否需要重新构建
$needBuild = $true
try {
    $imageExists = docker images chat-zinx-server:latest --format "{{.Repository}}"
    if ($imageExists) {
        Write-Host "聊天服务器镜像已存在" -ForegroundColor Gray
        $needBuild = $false
    }
} catch {
    # 镜像不存在，需要构建
}

if ($needBuild) {
    Write-Host "构建聊天服务器镜像..." -ForegroundColor Yellow
    
    # 创建 Dockerfile
    @"
FROM golang:1.22.6-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o chat-server .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/chat-server .
COPY --from=builder /app/conf ./conf

EXPOSE 9000 8080

CMD ["./chat-server"]
"@ | Out-File -FilePath "Dockerfile" -Encoding UTF8

    docker build -t chat-zinx-server:latest .
    Write-Host "✓ 聊天服务器镜像构建完成" -ForegroundColor Green
}

Set-Location $ProjectRoot

# 第四步：启动多个聊天服务器实例
Write-Host "`n=== 第四步：启动聊天服务器实例 ===" -ForegroundColor Green

# 创建服务器实例配置
for ($i = 1; $i -le $ServerCount; $i++) {
    $tcpPort = 9000 + $i - 1
    $httpPort = 8080 + $i - 1
    $serverName = "chat-server-$i"
    
    Write-Host "启动 $serverName (TCP:$tcpPort, HTTP:$httpPort)..." -ForegroundColor Yellow
    
    # 启动聊天服务器容器
    docker run -d `
        --name $serverName `
        --network chat-network `
        -p "${tcpPort}:9000" `
        -p "${httpPort}:8080" `
        -e SERVER_HOST="0.0.0.0" `
        -e SERVER_PORT="9000" `
        -e HTTP_PORT="8080" `
        -e SERVER_ID="chat-server-${i}" `
        -e IS_DEV="true" `
        chat-zinx-server:latest
    
    Write-Host "✓ $serverName 启动完成" -ForegroundColor Green
}

# 等待服务器启动
Write-Host "等待聊天服务器启动完成..." -ForegroundColor Yellow
Start-Sleep 10

# 第五步：验证系统状态
Write-Host "`n=== 第五步：系统状态验证 ===" -ForegroundColor Green

Write-Host "验证聊天服务器状态..." -ForegroundColor Yellow
for ($i = 1; $i -le $ServerCount; $i++) {
    $httpPort = 8080 + $i - 1
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:$httpPort/health" -TimeoutSec 5
        Write-Host "✓ Chat-Server-$i (HTTP:$httpPort) 运行正常" -ForegroundColor Green
    } catch {
        Write-Host "✗ Chat-Server-$i (HTTP:$httpPort) 状态异常" -ForegroundColor Red
    }
}

# 检查Consul中的服务注册
Write-Host "`n检查服务注册状态..." -ForegroundColor Yellow
try {
    $services = Invoke-RestMethod -Uri "http://localhost:8500/v1/catalog/service/chat-server" -TimeoutSec 5
    $serviceCount = $services.Count
    Write-Host "✓ Consul中已注册 $serviceCount 个聊天服务器实例" -ForegroundColor Green
    
    foreach ($service in $services) {
        Write-Host "  - $($service.ServiceID) @ $($service.ServiceAddress):$($service.ServicePort)" -ForegroundColor Gray
    }
} catch {
    Write-Host "✗ 无法获取服务注册信息" -ForegroundColor Red
}

# 显示系统访问信息
Write-Host "`n=== 系统访问信息 ===" -ForegroundColor Green
Write-Host "Consul UI: http://localhost:8500" -ForegroundColor Cyan
Write-Host "NATS监控: http://localhost:8222" -ForegroundColor Cyan
Write-Host "Grafana: http://localhost:3000 (admin/admin)" -ForegroundColor Cyan
Write-Host "Prometheus: http://localhost:9090" -ForegroundColor Cyan

Write-Host "`n聊天服务器实例:" -ForegroundColor Green
for ($i = 1; $i -le $ServerCount; $i++) {
    $tcpPort = 9000 + $i - 1
    $httpPort = 8080 + $i - 1
    Write-Host "  Chat-Server-$i: TCP:$tcpPort, HTTP:$httpPort" -ForegroundColor Cyan
}

# 显示测试命令
Write-Host "`n=== 测试命令 ===" -ForegroundColor Green
Write-Host "测试跨服务器通信:" -ForegroundColor Yellow
Write-Host "  1. 客户端A连接到 localhost:9000" -ForegroundColor Gray
Write-Host "  2. 客户端B连接到 localhost:9001" -ForegroundColor Gray  
Write-Host "  3. 客户端A向客户端B发送消息" -ForegroundColor Gray
Write-Host "  4. 观察消息是否通过NATS成功路由" -ForegroundColor Gray

if ($Logs) {
    Write-Host "`n=== 实时日志 ===" -ForegroundColor Green
    Write-Host "显示所有容器日志 (Ctrl+C 退出)..." -ForegroundColor Yellow
    docker-compose logs -f
} else {
    Write-Host "`n使用 -Logs 参数查看实时日志" -ForegroundColor Yellow
}

Write-Host "`n=== 分布式聊天系统启动完成! ===" -ForegroundColor Green
Write-Host "系统现在支持:" -ForegroundColor Yellow
Write-Host "  ✓ 跨服务器P2P消息传递" -ForegroundColor Green
Write-Host "  ✓ 分布式群组消息广播" -ForegroundColor Green  
Write-Host "  ✓ 服务自动发现与健康检查" -ForegroundColor Green
Write-Host "  ✓ 高可用性和故障转移" -ForegroundColor Green
Write-Host "  ✓ 水平扩展支持" -ForegroundColor Green 