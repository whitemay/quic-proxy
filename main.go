package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/quic-go/quic-go"
)

const (
	quicAddr = "localhost:4242"
	tcpAddr  = "localhost:8080"
)

// handleQUICSession 处理QUIC会话，将QUIC流的数据转发到TCP连接
func handleQUICSession(session quic.Connection) {
	fmt.Printf("新的QUIC会话来自 %s\n", session.RemoteAddr())

	// 从QUIC会话中接受一个流
	stream, err := session.AcceptStream(context.Background())
	if err != nil {
		log.Printf("接受流时出错: %s\n", err)
		return
	}

	// 连接到TCP服务器
	tcpConn, err := net.Dial("tcp", tcpAddr)
	if err != nil {
		log.Printf("连接到TCP服务器时出错: %s\n", err)
		// 关闭QUIC流，促使对方重新连接
		stream.Close()
		return
	}
	defer tcpConn.Close()

	fmt.Printf("已连接到TCP服务器 %s\n", tcpAddr)

	// 将QUIC流的数据转发到TCP连接
	go func() {
		_, err := io.Copy(tcpConn, stream)
		if err != nil {
			log.Printf("从QUIC复制数据到TCP时出错: %s\n", err)
		}
	}()

	// 将TCP连接的数据转发到QUIC流
	_, err = io.Copy(stream, tcpConn)
	if err != nil {
		log.Printf("从TCP复制数据到QUIC时出错: %s\n", err)
	}

	// 确保在函数结束时关闭QUIC流
	defer stream.Close()
}

func main() {
	// 从环境变量中获取证书文件地址
	certFile := os.Getenv("QUIC_CERT_FILE")
	keyFile := os.Getenv("QUIC_KEY_FILE")
	if certFile == "" || keyFile == "" {
		log.Fatalf("QUIC_CERT_FILE 和 QUIC_KEY_FILE 环境变量必须设置\n")
	}

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

	// 创建QUIC监听器
	listener, err := quic.ListenAddr(quicAddr, tlsConfig, nil)
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
		go handleQUICSession(session)
	}
}