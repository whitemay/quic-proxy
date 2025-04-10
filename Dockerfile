# 第一阶段：构建阶段
FROM golang:alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的依赖
RUN apk add --no-cache git build-base

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制项目代码
COPY . .

# 编译项目为静态链接的可执行文件
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o quic-proxy .

# 第二阶段：运行阶段
FROM alpine:latest

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制可执行文件
COPY --from=builder /app/quic-proxy .

# 暴露端口
EXPOSE 5100

# 启动命令
CMD ["./quic-proxy"]