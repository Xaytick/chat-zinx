module github.com/Xaytick/chat-zinx/chat-server

go 1.24.1

require github.com/Xaytick/zinx v0.0.0-20250515135912-f7eedf30ce5f

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/go-sql-driver/mysql v1.9.2 // indirect
)

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/uuid v1.6.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/crypto v0.38.0
)

replace github.com/Xaytick/chat-zinx/chat-server => ../chat-server
