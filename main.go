package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/quic-go/quic-go"
)

const (
	tcpAddr = "127.0.0.1"
)

func main() {
	// 从环境变量中获取证书文件地址
	certFile := os.Getenv("QUIC_CERT_FILE")
	keyFile := os.Getenv("QUIC_KEY_FILE")
	if certFile == "" || keyFile == "" {
		log.Fatalf("QUIC_CERT_FILE 和 QUIC_KEY_FILE 环境变量必须设置\n")
	}

	// 从环境变量中获取端口号，默认为5100
	portStr := os.Getenv("QUIC_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 {
		port = 5100
	}

	quicAddr := fmt.Sprintf(":%d", port)
	tcpAddrWithPort := fmt.Sprintf("%s:%d", tcpAddr, port)

	// 加载服务器证书
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("加载X509密钥对时出错: %s\n", err)
	}

	// 创建一个基本的TLS配置
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"quic-echo-example"},
	}

	// 创建QUIC监听器配置
	quicConfig := &quic.Config{
		MaxIdleTimeout: 30 * time.Minute, // 设置会话空闲超时时间为30分钟
	}

	// 创建QUIC监听器
	listener, err := quic.ListenAddr(quicAddr, tlsConfig, quicConfig)
	if err != nil {
		log.Fatalf("创建QUIC监听器时出错: %s\n", err)
	}
	defer listener.Close()

	fmt.Printf("正在监听QUIC连接 %s\n", quicAddr)

	// 接受QUIC会话并处理
	for {
		session, err := listener.Accept(context.Background())
		if err != nil {
			log.Printf("接受QUIC会话时出错: %s\n", err)
			continue
		}
		go handleQUICSession(session, tcpAddrWithPort)
	}
}