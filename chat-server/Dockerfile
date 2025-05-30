# 构建阶段
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o chat-server main.go

# 运行阶段
FROM alpine:latest

# 安装ca-certificates用于SSL连接
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/chat-server .

# 创建配置目录
RUN mkdir -p conf logs

# 设置权限
RUN chmod +x chat-server

# 暴露端口
EXPOSE 9000 8080

# 启动命令
CMD ["./chat-server"] 