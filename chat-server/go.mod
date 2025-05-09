module github.com/Xaytick/chat-zinx/chat-server

go 1.24.1

require github.com/Xaytick/zinx v0.0.0-20250508173059-940fe4b8a9e3

require (
	github.com/google/uuid v1.6.0
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/crypto v0.38.0
)

replace github.com/Xaytick/chat-zinx/chat-server => ../chat-server
