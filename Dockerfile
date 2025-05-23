# 第一阶段：构建阶段
FROM rust:1.87-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的构建工具
RUN apk add --no-cache musl-dev openssl-dev

# 复制项目代码（包括 Cargo.toml 和 Cargo.lock）
COPY . .

# 下载依赖
RUN cargo fetch

# 编译项目为静态链接的可执行文件
RUN cargo build --release --target x86_64-unknown-linux-musl

# 第二阶段：运行阶段
FROM alpine:3.18

# 设置工作目录
WORKDIR /app/

# 从构建阶段复制可执行文件
COPY --from=builder /app/target/x86_64-unknown-linux-musl/release/quic-proxy .

# 暴露端口
EXPOSE 5100

# 启动命令
CMD ["./quic-proxy"]