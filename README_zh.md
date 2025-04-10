# QUIC 代理

一个从 QUIC 到 TCP 的桥接工具。

## 概述

本项目提供了一个代理服务器，监听 QUIC 连接并将流量转发到 TCP 服务。旨在促进基于 QUIC 的客户端与传统基于 TCP 的服务之间的通信。

如需英文文档，请查看 [README.md](README.md)。

## 功能

- 监听 QUIC 连接并将数据转发到 TCP 服务。
- 支持通过环境变量配置端口、证书文件等。
- 包含 Dockerfile，便于跨平台构建和部署。

## 前置条件

- Go 1.20 或更高版本
- QUIC 兼容的证书（例如，使用 OpenSSL 生成）

## 安装

1. 克隆仓库：
   ```bash
   git clone https://github.com/whitemay/quic-proxy.git
   cd quic-proxy
   ```
2. 尝试运行：
   参考.env.example在当前目录下创建.env文件，并编辑。

   然后尝试：
   ```bash
   go get
   go run .
   ```
3. 构建docker镜像：
   ```bash
   docker build -t quic-proxy .
   docker run -d -p 443:443 quic-proxy
   ```