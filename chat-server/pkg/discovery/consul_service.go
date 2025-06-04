package discovery

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/consul/api"
)

type ConsulService struct {
	client   *api.Client
	serverID string
	config   *ServiceConfig
}

type ServiceConfig struct {
	Name    string            `json:"name"`
	ID      string            `json:"id"`
	Address string            `json:"address"`
	Port    int               `json:"port"`
	Tags    []string          `json:"tags"`
	Meta    map[string]string `json:"meta"`
}

type UserOnlineInfo struct {
	ServerID  string `json:"server_id"`
	Timestamp int64  `json:"timestamp"`
}

func NewConsulService(consulAddr string, config *ServiceConfig) (*ConsulService, error) {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulAddr

	client, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	// 验证连接
	_, err = client.Status().Leader()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to consul: %w", err)
	}

	return &ConsulService{
		client:   client,
		serverID: config.ID,
		config:   config,
	}, nil
}

// 注册服务
func (cs *ConsulService) RegisterService() error {
	// 健康检查配置
	check := &api.AgentServiceCheck{
		TCP:                            fmt.Sprintf("%s:%d", cs.config.Address, cs.config.Port),
		Interval:                       "10s",
		Timeout:                        "3s",
		DeregisterCriticalServiceAfter: "30s",
	}

	// 服务注册配置
	service := &api.AgentServiceRegistration{
		ID:      cs.config.ID,
		Name:    cs.config.Name,
		Tags:    cs.config.Tags,
		Address: cs.config.Address,
		Port:    cs.config.Port,
		Meta:    cs.config.Meta,
		Check:   check,
	}

	// 注册服务
	if err := cs.client.Agent().ServiceRegister(service); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	log.Printf("Service %s registered successfully at %s:%d", cs.config.ID, cs.config.Address, cs.config.Port)
	return nil
}

// 注销服务
func (cs *ConsulService) DeregisterService() error {
	if err := cs.client.Agent().ServiceDeregister(cs.serverID); err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	log.Printf("Service %s deregistered successfully", cs.serverID)
	return nil
}

// 发现服务
func (cs *ConsulService) DiscoverServices(serviceName string) ([]*api.ServiceEntry, error) {
	services, _, err := cs.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}

	return services, nil
}

// 获取健康的服务实例
func (cs *ConsulService) GetHealthyInstances(serviceName string) ([]string, error) {
	services, err := cs.DiscoverServices(serviceName)
	if err != nil {
		return nil, err
	}

	var instances []string
	for _, service := range services {
		instance := fmt.Sprintf("%s:%d",
			service.Service.Address,
			service.Service.Port)
		instances = append(instances, instance)
	}

	return instances, nil
}

// 监听服务变更
func (cs *ConsulService) WatchServices(serviceName string, callback func([]string)) {
	go func() {
		lastInstances := []string{}

		for {
			instances, err := cs.GetHealthyInstances(serviceName)
			if err != nil {
				log.Printf("Error getting healthy instances: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			// 检查实例是否变化
			if !stringSliceEqual(instances, lastInstances) {
				log.Printf("Service instances changed for %s: %v", serviceName, instances)
				callback(instances)
				lastInstances = instances
			}

			time.Sleep(10 * time.Second)
		}
	}()
}

// 设置键值对
func (cs *ConsulService) SetKV(key, value string) error {
	kv := cs.client.KV()
	pair := &api.KVPair{
		Key:   key,
		Value: []byte(value),
	}

	_, err := kv.Put(pair, nil)
	if err != nil {
		return fmt.Errorf("failed to set KV %s: %w", key, err)
	}

	return nil
}

// 获取键值对
func (cs *ConsulService) GetKV(key string) (string, error) {
	kv := cs.client.KV()
	pair, _, err := kv.Get(key, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get KV %s: %w", key, err)
	}

	if pair == nil {
		return "", fmt.Errorf("key not found: %s", key)
	}

	return string(pair.Value), nil
}

// 删除键值对
func (cs *ConsulService) DeleteKV(key string) error {
	kv := cs.client.KV()
	_, err := kv.Delete(key, nil)
	if err != nil {
		return fmt.Errorf("failed to delete KV %s: %w", key, err)
	}
	return nil
}

// 设置用户在线状态
func (cs *ConsulService) SetUserOnline(userUUID string) error {
	key := fmt.Sprintf("users/online/%s", userUUID)
	userInfo := UserOnlineInfo{
		ServerID:  cs.serverID,
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(userInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal user info: %w", err)
	}

	return cs.SetKV(key, string(data))
}

// 设置用户离线状态
func (cs *ConsulService) SetUserOffline(userUUID string) error {
	key := fmt.Sprintf("users/online/%s", userUUID)
	return cs.DeleteKV(key)
}

// 获取用户所在服务器
func (cs *ConsulService) GetUserServer(userUUID string) (string, error) {
	key := fmt.Sprintf("users/online/%s", userUUID)
	value, err := cs.GetKV(key)
	if err != nil {
		return "", err
	}

	var userInfo UserOnlineInfo
	if err := json.Unmarshal([]byte(value), &userInfo); err != nil {
		return "", fmt.Errorf("failed to unmarshal user info: %w", err)
	}

	// 检查时间戳，如果超过5分钟认为过期
	if time.Now().Unix()-userInfo.Timestamp > 300 {
		cs.DeleteKV(key) // 清理过期数据
		return "", fmt.Errorf("user status expired")
	}

	return userInfo.ServerID, nil
}

// 获取所有在线用户
func (cs *ConsulService) GetAllOnlineUsers() (map[string]string, error) {
	kv := cs.client.KV()
	pairs, _, err := kv.List("users/online/", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list online users: %w", err)
	}

	result := make(map[string]string)
	now := time.Now().Unix()

	for _, pair := range pairs {
		var userInfo UserOnlineInfo
		if err := json.Unmarshal(pair.Value, &userInfo); err != nil {
			continue
		}

		// 检查是否过期
		if now-userInfo.Timestamp > 300 {
			// 异步清理过期数据
			go cs.DeleteKV(pair.Key)
			continue
		}

		// 提取用户UUID
		userUUID := pair.Key[len("users/online/"):]
		result[userUUID] = userInfo.ServerID
	}

	return result, nil
}

// 设置配置
func (cs *ConsulService) SetConfig(key string, config interface{}) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configKey := fmt.Sprintf("config/%s", key)
	return cs.SetKV(configKey, string(data))
}

// 获取配置
func (cs *ConsulService) GetConfig(key string, result interface{}) error {
	configKey := fmt.Sprintf("config/%s", key)
	value, err := cs.GetKV(configKey)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(value), result); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// 获取服务统计信息
func (cs *ConsulService) GetServiceStats() (map[string]interface{}, error) {
	services, err := cs.DiscoverServices(cs.config.Name)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_instances": len(services),
		"healthy_count":   0,
		"instances":       []map[string]interface{}{},
	}

	healthyCount := 0
	for _, service := range services {
		healthyCount++
		instance := map[string]interface{}{
			"id":      service.Service.ID,
			"address": service.Service.Address,
			"port":    service.Service.Port,
			"tags":    service.Service.Tags,
			"meta":    service.Service.Meta,
		}
		stats["instances"] = append(stats["instances"].([]map[string]interface{}), instance)
	}

	stats["healthy_count"] = healthyCount
	return stats, nil
}

// 健康检查
func (cs *ConsulService) HealthCheck() error {
	_, err := cs.client.Agent().Self()
	if err != nil {
		return fmt.Errorf("consul health check failed: %w", err)
	}
	return nil
}

// 字符串切片比较
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
