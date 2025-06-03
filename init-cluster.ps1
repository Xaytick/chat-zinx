# Redis Cluster Initialization Script
Write-Host "=== Redis Cluster Initialization ===" -ForegroundColor Green

# Stop and restart Redis cluster containers to reset
Write-Host "`n1. Resetting Redis cluster containers..." -ForegroundColor Yellow
docker-compose stop redis-master-1 redis-master-2 redis-master-3 redis-slave-1 redis-slave-2 redis-slave-3

# Remove existing containers to start fresh
Write-Host "`n2. Removing existing Redis containers..." -ForegroundColor Yellow
docker-compose rm -f redis-master-1 redis-master-2 redis-master-3 redis-slave-1 redis-slave-2 redis-slave-3

# Start Redis containers again
Write-Host "`n3. Starting Redis containers..." -ForegroundColor Yellow
docker-compose up -d redis-master-1 redis-master-2 redis-master-3 redis-slave-1 redis-slave-2 redis-slave-3

# Wait for containers to be ready
Write-Host "`n4. Waiting for Redis nodes to start..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# Test all nodes are responding
Write-Host "`n5. Testing node connectivity..." -ForegroundColor Yellow
$nodes = @("redis-master-1", "redis-master-2", "redis-master-3", "redis-slave-1", "redis-slave-2", "redis-slave-3")
$ports = @(7001, 7002, 7003, 7004, 7005, 7006)

for ($i = 0; $i -lt $nodes.Length; $i++) {
    $node = $nodes[$i]
    $port = $ports[$i]
    try {
        $ping = docker exec $node redis-cli -p $port ping 2>$null
        if ($ping -eq "PONG") {
            Write-Host "[OK] $node responding" -ForegroundColor Green
        } else {
            Write-Host "[ERROR] $node not responding" -ForegroundColor Red
            exit 1
        }
    } catch {
        Write-Host "[ERROR] Cannot connect to $node" -ForegroundColor Red
        exit 1
    }
}

# Initialize cluster
Write-Host "`n6. Initializing Redis cluster..." -ForegroundColor Yellow
Write-Host "Command: redis-cli --cluster create redis-master-1:7001 redis-master-2:7002 redis-master-3:7003 redis-slave-1:7004 redis-slave-2:7005 redis-slave-3:7006 --cluster-replicas 1 --cluster-yes" -ForegroundColor Cyan

$initResult = docker exec redis-master-1 redis-cli --cluster create redis-master-1:7001 redis-master-2:7002 redis-master-3:7003 redis-slave-1:7004 redis-slave-2:7005 redis-slave-3:7006 --cluster-replicas 1 --cluster-yes 2>&1

Write-Host "`nCluster initialization output:" -ForegroundColor Yellow
Write-Host $initResult -ForegroundColor White

# Verify cluster status
Write-Host "`n7. Verifying cluster status..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

try {
    $clusterInfo = docker exec redis-master-1 redis-cli -p 7001 cluster info 2>$null
    $clusterNodes = docker exec redis-master-1 redis-cli -p 7001 cluster nodes 2>$null
    
    Write-Host "`nCluster Info:" -ForegroundColor Cyan
    Write-Host $clusterInfo -ForegroundColor White
    
    Write-Host "`nCluster Nodes:" -ForegroundColor Cyan
    Write-Host $clusterNodes -ForegroundColor White
    
    # Check if cluster is in OK state
    if ($clusterInfo -match "cluster_state:ok") {
        Write-Host "`n[SUCCESS] Redis cluster initialized successfully!" -ForegroundColor Green
        
        # Test cluster functionality
        Write-Host "`n8. Testing cluster functionality..." -ForegroundColor Yellow
        $setResult = docker exec redis-master-1 redis-cli -c -p 7001 set test-key "cluster-test" 2>$null
        $getResult = docker exec redis-master-1 redis-cli -c -p 7001 get test-key 2>$null
        
        if ($setResult -eq "OK" -and $getResult -eq "cluster-test") {
            Write-Host "[OK] Cluster read/write test successful" -ForegroundColor Green
            docker exec redis-master-1 redis-cli -c -p 7001 del test-key 2>$null | Out-Null
        } else {
            Write-Host "[WARN] Cluster read/write test failed" -ForegroundColor Yellow
        }
    } else {
        Write-Host "`n[ERROR] Cluster not in OK state" -ForegroundColor Red
    }
} catch {
    Write-Host "[ERROR] Failed to verify cluster status" -ForegroundColor Red
}

Write-Host "`n=== Initialization Complete ===" -ForegroundColor Green 