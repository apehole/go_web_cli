/**
 *@Author: IronHuang
 *@Date: 2020/8/18 9:06 下午
**/

package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"go_web_cli/dao/mysql"
	"go_web_cli/dao/redis"
	"go_web_cli/logger"
	"go_web_cli/routers"
	"go_web_cli/settings"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 1. 加载配置
	if err := settings.Init(); err != nil {
		fmt.Printf("init settings failed:%v\n", err)
		return
	}
	// 2. 初始化日志
	if err := logger.Init(settings.Conf.LogConfig); err != nil {
		fmt.Printf("init logger failed:%v\n", err)
		return
	}
	defer zap.L().Sync()
	zap.L().Debug("logger init success...")
	// 3. 初始化mysql
	if err := mysql.Init(settings.Conf.MysqlConfig); err != nil {
		fmt.Printf("init mysql failed:%v\n", err)
		return
	}
	defer mysql.Close()
	zap.L().Debug("mysql init success...")
	// 4. 初始化redis
	if err := redis.Init(settings.Conf.RedisConfig); err != nil {
		fmt.Printf("init redis failed:%v\n", err)
		return
	}
	defer redis.Close()
	zap.L().Debug("redis init success...")
	// 5. 注册路由
	r := routers.SetUp()
	zap.L().Debug("routers init success...")
	// 6. 启动服务（优雅关机）
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", settings.Conf.Port),
		Handler: r,
	}
	// 开启一个goroutine启动服务
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("listen:%s\n", zap.Error(err))
		}
	}()
	// 等待终端信号来优雅的关闭服务器，为关闭服务器操作设置一个五秒的超时
	quit := make(chan os.Signal, 1) // 创建一个接受信号的通道
	/*
		kill 默认会发送syscall.SIGTERM 信号
		kill -2 发送syscall.SIGINT 信号，我们常用的ctrl+c就是触发系统SIGINT信号
		kill -9 发送syscall.SIGKILL信号，但是不能被捕获，所以不需要添加它
		signal.Notify把收到的信号转发给quit
	*/
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit // 阻塞，当接受到上述两种信号时才会往下执行
	zap.L().Info("Shutdown Server...")
	fmt.Println("Shutdown Server...")
	// 创建一个5秒超时的context
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	// 5秒内优雅关闭服务器（将未处理完的请求处理完再关闭服务），超过五秒超时退出
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server Shutdown:", zap.Error(err))
	}
	time.Sleep(5 * time.Second)
	zap.L().Info("Server exiting")
	fmt.Println("Server exiting")
}
