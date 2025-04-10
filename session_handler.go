package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/quic-go/quic-go"
)

// handleQUICSession 处理QUIC会话，将QUIC流的数据转发到TCP连接
func handleQUICSession(session quic.Connection, tcpAddr string) {
	fmt.Printf("新的QUIC会话来自 %s\n", session.RemoteAddr())

	var streamCount int
	var streamCountMutex sync.Mutex

	for {
		// 接受QUIC会话中的流
		stream, err := session.AcceptStream(context.Background())
		if err != nil {
			// 如果会话被关闭，err 通常是 quic.ErrSessionClosed
			if err == quic.ErrTransportClosed {
				log.Printf("QUIC会话已关闭\n")
				return
			}
			log.Printf("接受流时出错: %s\n", err)
			continue
		}
		// 增加流计数
		streamCountMutex.Lock()
		streamCount++
		streamCountMutex.Unlock()

		go func(stream quic.Stream) {
			defer func() {
				// 减少流计数
				streamCountMutex.Lock()
				streamCount--
				streamCountMutex.Unlock()
			}()
			handleQUICStream(stream, tcpAddr)
		}(stream)
	}
}