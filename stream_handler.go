package main

import (
	"io"
	"log"
	"net"
)

// handleStream 处理单个流，将流的数据转发到 TCP 连接
func handleStream(stream io.ReadWriteCloser, tcpAddr string) {
	log.Printf("开始处理新的流\n")
	defer func() {
		log.Printf("关闭流\n")
		stream.Close()
	}()

	// 连接到 TCP 服务器
	tcpConn, err := net.Dial("tcp", tcpAddr)
	if err != nil {
		log.Printf("连接到 TCP 服务器时出错: %s\n", err)
		return
	}
	defer func() {
		log.Printf("关闭 TCP 连接\n")
		tcpConn.Close()
	}()
	log.Printf("已连接到 TCP 服务器 %s\n", tcpAddr)

	// 创建一个通道用于同步两个方向的转发完成状态
	done := make(chan struct{}, 2)

	// 将流的数据转发到 TCP 连接
	go func() {
		defer func() {
			log.Printf("完成从流到 TCP 的数据转发\n")
			done <- struct{}{}
		}()
		_, err := io.Copy(tcpConn, stream)
		if err != nil {
			if err == io.EOF {
				log.Printf("流已关闭\n")
			} else {
				log.Printf("从流复制数据到 TCP 时出错: %s\n", err)
			}
		}
	}()

	// 将 TCP 连接的数据转发到流
	go func() {
		defer func() {
			log.Printf("完成从 TCP 到流的数据转发\n")
			done <- struct{}{}
		}()
		_, err := io.Copy(stream, tcpConn)
		if err != nil {
			if err == io.EOF {
				log.Printf("TCP 连接已关闭\n")
			} else {
				log.Printf("从 TCP 复制数据到流时出错: %s\n", err)
			}
		}
	}()

	// 等待两个方向的转发完成
	<-done
	<-done
}
