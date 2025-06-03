# PowerShell Startup Script - Windows Version
# Redis Cluster Chat System Deployment

Write-Host "=== Chat System Redis Cluster Startup Script (Windows) ===" -ForegroundColor Green

# Check Docker availability
Write-Host "1. Checking Docker environment..." -ForegroundColor Yellow
try {
    $dockerVersion = docker --version
    Write-Host "[OK] Docker installed: $dockerVersion" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] Docker not installed or unavailable. Please install Docker Desktop" -ForegroundColor Red
    exit 1
}

try {
    $dockerComposeVersion = docker-compose --version
    Write-Host "[OK] Docker Compose installed: $dockerComposeVersion" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] Docker Compose unavailable" -ForegroundColor Red
    exit 1
}

# Create network if not exists
Write-Host "`n2. Creating Docker network..." -ForegroundColor Yellow
try {
    docker network create chat-network 2>$null
    Write-Host "[OK] Docker network chat-network created" -ForegroundColor Green
} catch {
    Write-Host "[INFO] Network chat-network already exists, skipping" -ForegroundColor Cyan
}

# Generate Redis cluster config files
Write-Host "`n3. Checking Redis cluster configuration files..." -ForegroundColor Yellow
Push-Location infrastructure/redis-cluster
if (Test-Path "scripts/create-configs.sh") {
    Write-Host "[INFO] Found create-configs.sh, may need to run in Git Bash or WSL" -ForegroundColor Cyan
    Write-Host "  Or check if redis-configs directory already has config files" -ForegroundColor Cyan
} 

# Check if config files exist
if (Test-Path "redis-configs") {
    $configFiles = Get-ChildItem "redis-configs" -Filter "*.conf"
    if ($configFiles.Count -ge 6) {
        Write-Host "[OK] Redis config files exist ($($configFiles.Count) files)" -ForegroundColor Green
    } else {
        Write-Host "[WARN] Redis config files incomplete, may need manual generation" -ForegroundColor Yellow
    }
} else {
    Write-Host "[WARN] redis-configs directory not found, will use default config" -ForegroundColor Yellow
}
Pop-Location

# Start base services (MySQL + Redis Cluster)
Write-Host "`n4. Starting MySQL and Redis Cluster services..." -ForegroundColor Yellow
Write-Host "Starting MySQL services..." -ForegroundColor Cyan
docker-compose up -d mysql-master mysql-slave1 mysql-slave2 mysql-shard0 mysql-shard1

Write-Host "Starting Redis Cluster nodes..." -ForegroundColor Cyan
docker-compose up -d redis-master-1 redis-master-2 redis-master-3 redis-slave-1 redis-slave-2 redis-slave-3

# Wait for Redis nodes to start
Write-Host "`n5. Waiting for Redis nodes to fully start..." -ForegroundColor Yellow
Start-Sleep -Seconds 15

# Initialize Redis cluster
Write-Host "`n6. Initializing Redis cluster..." -ForegroundColor Yellow
docker-compose up -d redis-cluster-init

# Wait for cluster initialization
Write-Host "`n7. Waiting for Redis cluster initialization..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# Verify cluster status
Write-Host "`n8. Verifying Redis cluster status..." -ForegroundColor Yellow
try {
    $clusterInfo = docker exec redis-master-1 redis-cli -p 7001 cluster nodes 2>$null
    if ($clusterInfo) {
        Write-Host "[OK] Redis Cluster node information:" -ForegroundColor Green
        Write-Host $clusterInfo -ForegroundColor White
    } else {
        Write-Host "[WARN] Cannot get cluster status, cluster may still be initializing" -ForegroundColor Yellow
    }
} catch {
    Write-Host "[WARN] Cluster status check failed, please check manually later" -ForegroundColor Yellow
}

# Start chat server
Write-Host "`n9. Starting chat server..." -ForegroundColor Yellow
docker-compose up -d chat-server

# Start monitoring services
Write-Host "`n10. Starting monitoring services..." -ForegroundColor Yellow
docker-compose up -d adminer prometheus grafana

Write-Host "`n=== Deployment Complete ===" -ForegroundColor Green

Write-Host "`nService Access Addresses:" -ForegroundColor Cyan
Write-Host "- Chat Server: localhost:9000 (TCP), localhost:8080 (HTTP)" -ForegroundColor White
Write-Host "- Database Admin: http://localhost:8081" -ForegroundColor White
Write-Host "- Prometheus: http://localhost:9090" -ForegroundColor White
Write-Host "- Grafana: http://localhost:3000 (admin/admin)" -ForegroundColor White

Write-Host "`nRedis Cluster Nodes:" -ForegroundColor Cyan
Write-Host "- Master 1: localhost:7001" -ForegroundColor White
Write-Host "- Master 2: localhost:7002" -ForegroundColor White
Write-Host "- Master 3: localhost:7003" -ForegroundColor White
Write-Host "- Slave 1: localhost:7004" -ForegroundColor White
Write-Host "- Slave 2: localhost:7005" -ForegroundColor White
Write-Host "- Slave 3: localhost:7006" -ForegroundColor White

Write-Host "`nUseful Commands:" -ForegroundColor Cyan
Write-Host "- Check status: .\scripts\deployment\check-cluster.ps1" -ForegroundColor White
Write-Host "- View logs: docker-compose logs chat-server" -ForegroundColor White
Write-Host "- Stop services: docker-compose down" -ForegroundColor White
Write-Host "- View all containers: docker ps" -ForegroundColor White

Write-Host "`nStartup complete! Please wait a few minutes for all services to fully initialize" -ForegroundColor Green 