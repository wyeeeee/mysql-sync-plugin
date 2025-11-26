package main

import (
	"log"
	"mysql-sync-plugin/config"
	"mysql-sync-plugin/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 设置Gin模式
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	r := gin.Default()

	// CORS中间件(允许钉钉调用)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Ding-Docs-Timestamp, Ding-Docs-Signature")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 创建处理器
	h := handler.NewHandler()

	// 健康检查(无需签名)
	r.GET("/health", h.Health)

	// 前端配置页面使用的API(无需签名,用于配置阶段)
	r.POST("/api/tables", h.GetTables)
	r.POST("/api/fields", h.GetFields)

	// AI表格服务端调用的API(需要签名验证)
	api := r.Group("/api")
	api.POST("/sheet_meta", h.SheetMeta)
	api.POST("/records", h.Records)

	// 启动服务
	addr := ":" + cfg.ServerPort
	log.Printf("MySQL同步插件服务启动在端口 %s", cfg.ServerPort)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
