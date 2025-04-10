package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/quic-go/quic-go"
)

// handleQUICStream 处理单个QUIC流，将流的数据转发到TCP连接
func handleQUICStream(stream quic.Stream, tcpAddr string) {
	defer stream.Close()

	// 连接到TCP服务器
	tcpConn, err := net.Dial("tcp", tcpAddr)
	if err != nil {
		log.Printf("连接到TCP服务器时出错: %s\n", err)
		// 关闭QUIC流，促使对方重新连接
		return
	}
	defer tcpConn.Close()

	fmt.Printf("已连接到TCP服务器 %s\n", tcpAddr)

	// 将QUIC流的数据转发到TCP连接
	go func() {
		_, err := io.Copy(tcpConn, stream)
		if err != nil {
			// 检查是否是因为流被关闭导致的错误
			if err == io.EOF {
				log.Printf("QUIC流已关闭\n")
			} else {
				log.Printf("从QUIC复制数据到TCP时出错: %s\n", err)
			}
		}
		// 关闭TCP连接，确保资源释放
		tcpConn.Close()
	}()

	// 将TCP连接的数据转发到QUIC流
	_, err = io.Copy(stream, tcpConn)
	if err != nil {
		log.Printf("从TCP复制数据到QUIC时出错: %s\n", err)
	}
}