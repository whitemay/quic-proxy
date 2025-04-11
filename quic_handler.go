package main

import (
	"context"
	"log"

	"github.com/quic-go/quic-go"
)

// handleQUICSession 处理 QUIC 会话，将 QUIC 流的数据转发到 TCP 连接
func handleQUICSession(connection quic.Connection, tcpAddr string) {
	log.Printf("新的 QUIC 会话来自 %s\n", connection.RemoteAddr())
	// defer func() {
	// 	log.Printf("关闭 QUIC 会话来自 %s\n", connection.RemoteAddr())
	// 	connection.CloseWithError(0x42, "I don't want to talk to you anymore 🙉")
	// }()

	for {
		// 接受 QUIC 会话中的流
		stream, err := connection.AcceptStream(context.Background())
		if err != nil {
			if err == quic.ErrServerClosed {
				log.Printf("QUIC 会话已关闭\n")
				return
			}
			log.Printf("接受流时出错: %s\n", err)
			continue
		}
		log.Printf("接受新的 QUIC 流来自 %s\n", connection.RemoteAddr())
		go handleStream(stream, tcpAddr)
	}
}
