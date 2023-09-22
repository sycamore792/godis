package server

import (
	"awesomeProject/lib/logger"
	server "awesomeProject/server/interface"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var ClientCounter int

// 配置对象
type Config struct {
	Address    string        `yaml:"address"`
	MaxConnect uint32        `yaml:"max-connect"`
	Timeout    time.Duration `yaml:"timeout"`
}

func ListenAndServeWithSignal(cfg *Config, handler server.Handler) error {
	// ListenAndServeWithSignal 监听中断信号并通过 closeChan 通知服务器关闭
	closeChan := make(chan struct{})
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("bind: %s, start listening...", cfg.Address))
	ListenerService(listener, handler, closeChan)
	return nil
}

func ListenerService(listener net.Listener, handler server.Handler, closeChan <-chan struct{}) {

	errCh := make(chan error, 1)
	defer close(errCh)
	go func() {
		select {
		case <-closeChan:
			logger.Info("收到退出信号")
		case er := <-errCh:
			logger.Info(fmt.Sprintf("接收到错误: %s", er.Error()))
		}
		logger.Info("准备关闭")
		_ = listener.Close() // 停止监听，listener.Accept()会立即返回 io.EOF
		_ = handler.Close()  // 关闭应用层服务器
	}()

	ctx := context.Background()
	var waitDone sync.WaitGroup
	for {
		// 监听端口, 阻塞直到收到新连接或者出现错误
		conn, err := listener.Accept()
		if err != nil {
			// learn from net/http/serve.go#Serve()
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				logger.Infof("accept occurs temporary error: %v, retry in 5ms", err)
				time.Sleep(5 * time.Millisecond)
				continue
			}
			errCh <- err
			break
		}
		logger.Info("接受新的连接")
		ClientCounter++
		waitDone.Add(1)
		// 开启 goroutine 来处理新连接
		go func() {
			defer func() {
				waitDone.Done()
				ClientCounter--
			}()
			handler.Handler(ctx, conn)
		}()
	}
	waitDone.Wait()
}
