package main

import (
	"io"
	"log"
	"net"
)

// handleStream 处理单个流，将流的数据转发到 TCP 连接
func handleStream(stream io.ReadWriteCloser, tcpAddr string) {
	defer stream.Close()

	// 连接到 TCP 服务器
	tcpConn, err := net.Dial("tcp", tcpAddr)
	if err != nil {
		log.Printf("连接到 TCP 服务器时出错: %s\n", err)
		return
	}
	defer tcpConn.Close()

	log.Printf("已连接到 TCP 服务器 %s\n", tcpAddr)

	// 将流的数据转发到 TCP 连接
	go func() {
		_, err := io.Copy(tcpConn, stream)
		if err != nil {
			if err == io.EOF {
				log.Printf("流已关闭\n")
			} else {
				log.Printf("从流复制数据到 TCP 时出错: %s\n", err)
			}
		}
		tcpConn.Close()
	}()

	// 将 TCP 连接的数据转发到流
	_, err = io.Copy(stream, tcpConn)
	if err != nil {
		log.Printf("从 TCP 复制数据到流时出错: %s\n", err)
	}
}
