package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"spider-go/internal/api"
	"spider-go/internal/app"
	"strconv"
	"syscall"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 创建依赖注入容器
	container, err := app.NewContainer("./config")
	if err != nil {
		log.Fatalf("初始化容器失败: %v", err)
	}
	defer func() {
		if err := container.Close(); err != nil {
			log.Printf("关闭资源失败: %v", err)
		}
	}()

	// 2. 创建 Gin 引擎
	r := gin.Default()

	// 3. 设置路由
	api.SetupRoutes(r, container)

	// 4. 启动服务器
	port := container.Config.App.Port
	addr := ":" + strconv.Itoa(port)

	log.Printf("服务器启动在端口: %d\n", port)

	// 5. 优雅关闭
	go func() {
		if err := r.Run(addr); err != nil {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n正在关闭服务器...")
}
