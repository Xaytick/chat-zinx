module github.com/Xaytick/chat-zinx/chat-server

go 1.24.1

require (
	github.com/Xaytick/zinx v0.0.0-20250518152801-3cac1b237c94
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/hashicorp/consul/api v1.32.1
	// 新增分布式相关依赖
	github.com/nats-io/nats.go v1.43.0
	gorm.io/driver/mysql v1.5.7
	gorm.io/gorm v1.25.11
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.5.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.37.0
	golang.org/x/exp v0.0.0-20250305212735-054e65f0b394 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
)

require github.com/google/uuid v1.6.0

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/pkg/errors v0.9.1 // indirect
)

replace github.com/Xaytick/chat-zinx/chat-server => ../chat-server
