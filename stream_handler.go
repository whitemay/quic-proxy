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
	// tcpConn.Write([]byte("hello"))

	// // 创建一个通道用于同步两个方向的转发完成状态
	done := make(chan struct{}, 2)

	// 将 TCP 连接的数据转发到流
	// 使用独立 Goroutine 处理此方向的数据流
	go func() {
		defer func() {
			log.Printf("完成从 TCP 到流的数据转发\n")
			done <- struct{}{}
			stream.Close() // 关闭流，触发另一方向的读取结束
		}()
		log.Printf("开始从 TCP 连接到流的数据转发\n")
		// _, err := io.Copy(stream, tcpConn)
		// if err != nil {
		// 	if err == io.EOF {
		// 		log.Printf("TCP 连接已关闭\n")
		// 	} else {
		// 		log.Printf("从 TCP 复制数据到流时出错: %s\n", err)
		// 	}
		// }
		buffer := make([]byte, 4096) // 调试阶段可使用较小的缓冲区
		for {
			n, readErr := tcpConn.Read(buffer)
			if readErr != nil {
				if readErr == io.EOF {
					log.Printf("TCP 连接读取结束，关闭流写入\n")
				} else {
					log.Printf("从 TCP 读取数据时出错: %v", readErr)
				}
				break
			}
			// log.Printf("从 TCP 读取 %d 字节: % X", n, buffer[:n]) // 打印十六进制数据
			_, writeErr := stream.Write(buffer[:n])
			if writeErr != nil {
				log.Printf("向流写入数据时出错: %v", writeErr)
				break
			}
			// log.Printf("成功向流写入 %d 字节", written)
		}
	}()

	// 将流的数据转发到 TCP 连接
	// 使用独立 Goroutine 处理此方向的数据流
	go func() {
		defer func() {
			log.Printf("完成从流到 TCP 的数据转发\n")
			done <- struct{}{}
			tcpConn.Close() // 关闭 TCP 连接，触发另一方向的读取结束
		}()
		log.Printf("开始从流到 TCP 的数据转发\n")
		// _, err := io.Copy(tcpConn, stream)
		// if err != nil {
		// 	if err == io.EOF {
		// 		log.Printf("流已关闭\n")
		// 	} else {
		// 		log.Printf("从流复制数据到 TCP 时出错: %s\n", err)
		// 	}
		// }
		buffer := make([]byte, 4096)
		for {
			n, readErr := stream.Read(buffer)
			if readErr != nil {
				if readErr == io.EOF {
					log.Printf("流读取结束，关闭 TCP 写入\n")
				} else {
					log.Printf("从流读取数据时出错: %v", readErr)
				}
				break
			}
			// log.Printf("从流读取 %d 字节: % X", n, buffer[:n]) // 打印十六进制数据
			_, writeErr := tcpConn.Write(buffer[:n])
			if writeErr != nil {
				log.Printf("向 TCP 写入数据时出错: %v", writeErr)
				break
			}
			// log.Printf("成功向 TCP 写入 %d 字节", written)
		}
	}()

	// 等待两个方向的转发完成
	// 确保资源被正确释放
	<-done
	<-done
}
