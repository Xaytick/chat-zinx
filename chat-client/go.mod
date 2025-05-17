module github.com/Xaytick/chat-zinx/chat-client

go 1.24.1

require github.com/Xaytick/chat-zinx/chat-server v0.0.0-20250515140247-bcf35e06d36a

require github.com/golang-jwt/jwt v3.2.2+incompatible // indirect

replace github.com/Xaytick/chat-zinx/chat-server => ../chat-server
