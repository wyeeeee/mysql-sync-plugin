package main

import (
	"log"
	"mysql-sync-plugin/auth"
	"mysql-sync-plugin/config"
	"mysql-sync-plugin/handler"
	"mysql-sync-plugin/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化日志存储
	if err := logger.GetStore().Init(cfg.DBPath); err != nil {
		log.Fatalf("初始化日志数据库失败: %v", err)
	}
	defer logger.GetStore().Close()

	// 初始化认证存储
	if err := auth.GetStore().Init(cfg.DBPath); err != nil {
		log.Fatalf("初始化认证数据库失败: %v", err)
	}
	defer auth.GetStore().Close()

	mainLog := logger.New("main")
	mainLog.Info("启动", "MySQL同步插件服务正在启动")

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
	adminH := handler.NewAdminHandler()
	authH := handler.NewAuthHandler()

	// 创建基础路径组 /data
	baseGroup := r.Group("/data")

	// 健康检查(无需签名)
	baseGroup.GET("/health", h.Health)

	// API路由组
	api := baseGroup.Group("/api")
	{
		// 前端配置页面使用的API(无需签名,用于配置阶段)
		api.POST("/databases", h.GetDatabases)
		api.POST("/tables", h.GetTables)
		api.POST("/fields", h.GetFields)
		api.POST("/preview_sql", h.PreviewSQL)

		// AI表格服务端调用的API(需要签名验证)
		api.POST("/sheet_meta", h.SheetMeta)
		api.POST("/records", h.Records)
	}

	// 管理后台认证API(无需登录)
	adminAuth := baseGroup.Group("/admin/api")
	{
		adminAuth.POST("/login", authH.Login)
	}

	// 管理后台API路由组(需要登录)
	adminAPI := baseGroup.Group("/admin/api")
	adminAPI.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	})
	adminAPI.Use(auth.AdminAuthMiddleware())
	{
		adminAPI.POST("/logout", authH.Logout)
		adminAPI.GET("/user/current", authH.GetCurrentUser)
		adminAPI.POST("/user/password", authH.ChangePassword)
		adminAPI.GET("/logs", adminH.GetLogs)
		adminAPI.GET("/logs/stats", adminH.GetLogStats)
		adminAPI.POST("/logs/clean", adminH.CleanLogs)
		adminAPI.GET("/system/info", adminH.GetSystemInfo)
	}

	// 静态文件服务 - 插件前端页面
	baseGroup.Static("/assets", "./static/assets")
	baseGroup.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})
	baseGroup.StaticFile("/favicon.ico", "./static/favicon.ico")

	// 管理后台静态文件服务（带缓存控制）
	baseGroup.GET("/admin/assets/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		c.Header("Cache-Control", "no-cache, must-revalidate")
		c.File("./admin/assets" + filepath)
	})
	baseGroup.GET("/admin", func(c *gin.Context) {
		c.Request.Header.Del("If-Modified-Since")
		c.Request.Header.Del("If-None-Match")
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.File("./admin/index.html")
	})

	// SPA 路由支持：处理前端路由路径
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 管理后台 SPA 路由（/data/admin 开头，但不是静态资源）
		if len(path) >= 11 && path[:11] == "/data/admin" {
			// 排除静态资源路径
			if len(path) > 18 && path[:18] == "/data/admin/assets" {
				c.Status(404)
				return
			}
			// 排除 API 路径
			if len(path) > 15 && path[:15] == "/data/admin/api" {
				c.Status(404)
				return
			}
			c.Request.Header.Del("If-Modified-Since")
			c.Request.Header.Del("If-None-Match")
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
			c.File("./admin/index.html")
			return
		}

		// 插件前端不需要 SPA 路由支持（它是配置页面，不是 SPA）
		c.Status(404)
	})

	// 启动服务
	addr := ":" + cfg.ServerPort
	mainLog.Infof("启动", "服务启动在端口 %s", cfg.ServerPort)
	mainLog.Info("启动", "插件前端: http://localhost:"+cfg.ServerPort+"/data")
	mainLog.Info("启动", "管理后台: http://localhost:"+cfg.ServerPort+"/data/admin")
	if err := r.Run(addr); err != nil {
		mainLog.Errorf("启动", "服务启动失败: %v", err)
		log.Fatalf("服务启动失败: %v", err)
	}
}
