# PowerShell Check Script - Windows Version
# Redis Cluster Status Check

Write-Host "=== Redis Cluster Status Check (Windows) ===" -ForegroundColor Green

# Check all Redis nodes
Write-Host "`n1. Checking Redis node status:" -ForegroundColor Yellow

$ports = @(7001, 7002, 7003, 7004, 7005, 7006)
$nodeTypes = @("master", "master", "master", "slave", "slave", "slave")
$nodeNumbers = @("1", "2", "3", "1", "2", "3")

for ($i = 0; $i -lt $ports.Length; $i++) {
    $port = $ports[$i]
    $nodeType = $nodeTypes[$i]
    $nodeNumber = $nodeNumbers[$i]
    $containerName = "redis-$nodeType-$nodeNumber"
    
    # Check if container is running
    $containerRunning = docker ps --filter "name=$containerName" --format "{{.Names}}" 2>$null
    
    if ($containerRunning -eq $containerName) {
        Write-Host "[OK] $containerName (port $port) - Running" -ForegroundColor Green
        
        # Test Redis service response
        try {
            $pingResult = docker exec $containerName redis-cli -p $port ping 2>$null
            if ($pingResult -eq "PONG") {
                Write-Host "  - Redis service responding normally" -ForegroundColor Cyan
            } else {
                Write-Host "  - Redis service not responding" -ForegroundColor Red
            }
        } catch {
            Write-Host "  - Redis service not responding" -ForegroundColor Red
        }
    } else {
        Write-Host "[ERROR] $containerName (port $port) - Not running" -ForegroundColor Red
    }
}

Write-Host "`n2. Checking cluster status:" -ForegroundColor Yellow

try {
    $clusterInfo = docker exec redis-master-1 redis-cli -p 7001 cluster info 2>$null
    $clusterState = ($clusterInfo | Select-String "cluster_state:ok")
    
    if ($clusterState) {
        Write-Host "[OK] Redis Cluster status normal" -ForegroundColor Green
        
        Write-Host "`n3. Cluster node information:" -ForegroundColor Yellow
        $clusterNodes = docker exec redis-master-1 redis-cli -p 7001 cluster nodes 2>$null
        if ($clusterNodes) {
            Write-Host $clusterNodes -ForegroundColor White
        }
        
        Write-Host "`n4. Cluster slot distribution:" -ForegroundColor Yellow
        $clusterSlots = docker exec redis-master-1 redis-cli -p 7001 cluster slots 2>$null
        if ($clusterSlots) {
            Write-Host $clusterSlots -ForegroundColor White
        }
        
        Write-Host "`n5. Testing cluster read/write:" -ForegroundColor Yellow
        
        # Test write
        try {
            $setResult = docker exec redis-master-1 redis-cli -c -p 7001 set test-key "Hello Redis Cluster" 2>$null
            if ($setResult -eq "OK") {
                Write-Host "[OK] Cluster write test successful" -ForegroundColor Green
                
                # Test read
                $getValue = docker exec redis-master-1 redis-cli -c -p 7001 get test-key 2>$null
                if ($getValue -eq "Hello Redis Cluster") {
                    Write-Host "[OK] Cluster read test successful" -ForegroundColor Green
                    # Clean up test data
                    docker exec redis-master-1 redis-cli -c -p 7001 del test-key 2>$null | Out-Null
                } else {
                    Write-Host "[ERROR] Cluster read test failed" -ForegroundColor Red
                }
            } else {
                Write-Host "[ERROR] Cluster write test failed" -ForegroundColor Red
            }
        } catch {
            Write-Host "[ERROR] Cluster write test failed" -ForegroundColor Red
        }
        
    } else {
        Write-Host "[ERROR] Redis Cluster status abnormal or not initialized" -ForegroundColor Red
        Write-Host "`nPossible causes:" -ForegroundColor Yellow
        Write-Host "  - Cluster not fully initialized" -ForegroundColor White
        Write-Host "  - Some nodes offline" -ForegroundColor White
        Write-Host "  - Network connection issues" -ForegroundColor White
        
        Write-Host "`nSuggested actions:" -ForegroundColor Yellow
        Write-Host "  - Wait a few minutes and retry" -ForegroundColor White
        Write-Host "  - Check Docker logs: docker-compose logs redis-cluster-init" -ForegroundColor White
        Write-Host "  - Manual cluster initialization:" -ForegroundColor White
        Write-Host "    docker exec redis-master-1 redis-cli --cluster create redis-master-1:7001 redis-master-2:7002 redis-master-3:7003 redis-slave-1:7004 redis-slave-2:7005 redis-slave-3:7006 --cluster-replicas 1 --cluster-yes" -ForegroundColor Gray
    }
} catch {
    Write-Host "[ERROR] Cannot connect to Redis cluster" -ForegroundColor Red
    Write-Host "Please check if redis-master-1 container is running" -ForegroundColor Yellow
}

Write-Host "`n6. Chat server connection status:" -ForegroundColor Yellow

# Check if chat server container is running
$chatServerRunning = docker ps --filter "name=chat-server" --format "{{.Names}}" 2>$null

if ($chatServerRunning -eq "chat-server") {
    Write-Host "[OK] Chat server container is running" -ForegroundColor Green
    
    # Try health check (if health check endpoint exists)
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -TimeoutSec 5 -UseBasicParsing 2>$null
        if ($response.StatusCode -eq 200) {
            Write-Host "[OK] Chat server health check passed" -ForegroundColor Green
        } else {
            Write-Host "[WARN] Chat server response abnormal (status code: $($response.StatusCode))" -ForegroundColor Yellow
        }
    } catch {
        Write-Host "[WARN] Chat server health check failed or no health check endpoint available" -ForegroundColor Yellow
        Write-Host "  Try: docker-compose logs chat-server" -ForegroundColor Cyan
    }
} else {
    Write-Host "[ERROR] Chat server container not running" -ForegroundColor Red
}

Write-Host "`n7. Service port check:" -ForegroundColor Yellow

$portsToCheck = @(
    @{Port=3316; Service="MySQL Master"},
    @{Port=7001; Service="Redis Master 1"},
    @{Port=7002; Service="Redis Master 2"},
    @{Port=7003; Service="Redis Master 3"},
    @{Port=7004; Service="Redis Slave 1"},
    @{Port=7005; Service="Redis Slave 2"},
    @{Port=7006; Service="Redis Slave 3"},
    @{Port=9000; Service="Chat Server TCP"},
    @{Port=8080; Service="Chat Server HTTP"}
)

foreach ($portInfo in $portsToCheck) {
    try {
        $connection = Test-NetConnection -ComputerName localhost -Port $portInfo.Port -InformationLevel Quiet -WarningAction SilentlyContinue
        if ($connection) {
            Write-Host "[OK] $($portInfo.Service) (port $($portInfo.Port)) - Accessible" -ForegroundColor Green
        } else {
            Write-Host "[ERROR] $($portInfo.Service) (port $($portInfo.Port)) - Not accessible" -ForegroundColor Red
        }
    } catch {
        Write-Host "[WARN] $($portInfo.Service) (port $($portInfo.Port)) - Check failed" -ForegroundColor Yellow
    }
}

Write-Host "`n=== Useful commands ===" -ForegroundColor Cyan
Write-Host "- View all containers: docker ps" -ForegroundColor White
Write-Host "- View service logs: docker-compose logs [service-name]" -ForegroundColor White
Write-Host "- Restart service: docker-compose restart [service-name]" -ForegroundColor White
Write-Host "- Connect to Redis: docker exec -it redis-master-1 redis-cli -c -p 7001" -ForegroundColor White
Write-Host "- Stop all services: docker-compose down" -ForegroundColor White

Write-Host "`n=== Check Complete ===" -ForegroundColor Green 