package server

import (
	"awesomeProject/lib/logger"
	"awesomeProject/lib/sync/atomic"
	"awesomeProject/lib/sync/wait"
	"bufio"
	"context"
	"io"
	"net"
	"sync"
	"time"
)

// 记录所有的活跃的链接 以及自身的状态
type AliveConnHandler struct {
	activateConn sync.Map
	closing      atomic.Boolean
}

func GetAHandler() *AliveConnHandler {
	return &AliveConnHandler{}
}

// 客户端与服务器的链接 waitgroup表示有多少个程序正在并发
type Client struct {
	Conn    net.Conn
	Waiting wait.WaitingGroup
}

func (h *AliveConnHandler) Handle(ctx context.Context, conn net.Conn) {

	// 关闭中的 handler 不会处理新连接
	if h.closing.Get() {
		conn.Close()
		return
	}

	//初始化客户端
	client := &Client{
		Conn: conn,
	}
	//存储存活的客户端
	h.activateConn.Store(client, struct{}{})

	// 使用bufio标准库提供的缓冲区功能 读连接的内容
	reader := bufio.NewReader(conn)
	for {
		// ReadString 会一直阻塞直到遇到分隔符'\n'
		// 遇到分隔符后会返回与上次的分隔符之间的数据  包括最后的分隔符本身
		//	如果读取异常 则处理
		msg, err := reader.ReadString('\n')
		// 处理错误
		if err != nil {
			if err == io.EOF {
				logger.Info("connection closed")
				h.activateConn.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}
		// 任务开始 wg++
		client.Waiting.Add(1)

		// 模拟关闭时未完成发送的情况
		//logger.Info("sleeping")
		//time.Sleep(10 * time.Second)

		b := []byte(msg)
		conn.Write(b)
		// 发送完毕, 结束waiting
		client.Waiting.Done()
	}
}

// 关闭客户端连接
func (c *Client) Close() error {
	// 等待数据发送完成或超时
	c.Waiting.WaitWithTimeout(10 * time.Second)
	c.Conn.Close()
	return nil
}

// 关闭服务器
func (h *AliveConnHandler) Close() error {
	logger.Info("handler shutting down...")
	h.closing.Set(true)
	// 逐个关闭客户端连接
	h.activateConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*Client)
		client.Close()
		return true
	})
	return nil
}

//func (h *AliveConnHandler) Close() error {
//	logger.Info("handler shutting down...")
//	// 关闭
//	h.closing.Set(true)
//	h.activateConn.Range(func(key, value interface{}) bool {
//		// 遍历所有的客户端 一个个关闭
//		c := key.(*Client)
//		c.Close()
//		// return true进行下一次遍历
//		return true
//	})
//	return nil
//}
