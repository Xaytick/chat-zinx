module github.com/Xaytick/chat-zinx/chat-server

go 1.24.1

require github.com/Xaytick/zinx v0.0.0-20250515135912-f7eedf30ce5f

require golang.org/x/text v0.25.0 // indirect

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-sql-driver/mysql v1.7.1 // indirect; GORM 推荐的 mysql driver 版本，或者你可以使用 v1.8.x 或更新的
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	gorm.io/driver/mysql v1.5.7 // GORM MySQL driver
	gorm.io/gorm v1.25.11 // GORM 本身
)

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/uuid v1.6.0
	// github.com/jmoiron/sqlx v1.4.0 // 移除 sqlx
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/crypto v0.38.0
)

replace github.com/Xaytick/chat-zinx/chat-server => ../chat-server
