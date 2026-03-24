// main.go
package main

import (
	"blog/config"
	"blog/models"
	"blog/router"
	"log"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 初始化数据库
	models.InitDB(cfg)

	// 设置路由
	r := router.SetupRouter(cfg)

	// 打印所有路由（调试用）
	log.Println("Registered routes:")
	for _, route := range r.Routes() {
		log.Printf("  %-6s %s", route.Method, route.Path)
	}

	// 启动服务器
	log.Printf("Server is running on http://localhost:%s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
