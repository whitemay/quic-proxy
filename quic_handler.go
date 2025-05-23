package main

import (
	"context"
	"errors"
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
	defer func() {
		log.Printf("关闭 QUIC 会话来自 %s\n", connection.RemoteAddr())
		connection.CloseWithError(0x42, "I don't want to talk to you anymore 🙉")
	}()

	for {
		// 接受 QUIC 会话中的流
		stream, err := connection.AcceptStream(context.Background())
		if err != nil {
			var (
				statelessResetErr   *quic.StatelessResetError
				handshakeTimeoutErr *quic.HandshakeTimeoutError
				idleTimeoutErr      *quic.IdleTimeoutError
				appErr              *quic.ApplicationError
				transportErr        *quic.TransportError
				vnErr               *quic.VersionNegotiationError
			)
			switch {
			case errors.As(err, &statelessResetErr):
				log.Println("无状态重置")
			case errors.As(err, &handshakeTimeoutErr):
				log.Println("QUIC 会话因握手超时关闭")
			case errors.As(err, &idleTimeoutErr):
				log.Printf("QUIC 会话因空闲超时关闭\n")
			case errors.As(err, &appErr):
				// application error
				remote := appErr.Remote // was the error triggered by the peer?
				var closer string
				if remote {
					closer = "远程"
				} else {
					closer = "本地"
				}
				errorCode := appErr.ErrorCode       // application-defined error code
				errorMessage := appErr.ErrorMessage // application-defined error message
				log.Printf("QUIC 会话因应用错误被 %s 关闭: %s (类型: %T)\n", closer, errorMessage, errorCode)
			case errors.As(err, &transportErr):
				// transport error
				var closer string
				if transportErr.Remote {
					closer = "远程"
				} else {
					closer = "本地"
				}
				errorCode := transportErr.ErrorCode       // error code (RFC 9000, section 20.1)
				errorMessage := transportErr.ErrorMessage // error message
				log.Printf("QUIC 会话因传输错误被 %s 关闭: %s (类型: %T)\n", closer, errorMessage, errorCode)
			case errors.As(err, &vnErr):
				// version negotation error
				ourVersions := vnErr.Ours     // locally supported QUIC versions
				theirVersions := vnErr.Theirs // QUIC versions support by the remote
				log.Printf("QUIC 会话因版本协商错误关闭: %s (我的: %s)\n", theirVersions, ourVersions)
			}
			// 任何错误都应该停止连接的处理过程
			return
		}
		log.Printf("接受新的 QUIC 流来自 %s\n", connection.RemoteAddr())
		go handleStream(stream, tcpAddr)
	}
}
