# 第一阶段：构建阶段
FROM golang:alpine AS builder

# 设置工作目录
# WORKDIR /app
ENV PATH=/root/go/bin:${PATH}

# 安装必要的依赖
RUN apk add --no-cache git
RUN go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 xcaddy build --with github.com/mholt/caddy-l4

# 复制go.mod和go.sum文件
# COPY go.mod go.sum ./

# 下载依赖
# RUN go mod download

# 复制项目代码
# COPY . .

# 编译项目为静态链接的可执行文件
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o quic-proxy .

# 第二阶段：运行阶段
FROM alpine:latest

# 设置工作目录
WORKDIR /app/

# 从构建阶段复制可执行文件
COPY --from=builder /go/caddy /app/caddy

# 暴露端口
EXPOSE 5100

# 启动命令
ENTRYPOINT [ "/app/caddy" ]
CMD ["run", "--config", "/etc/Caddyfile"]