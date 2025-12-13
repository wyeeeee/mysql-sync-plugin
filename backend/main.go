package main

import (
	"log"
	"mysql-sync-plugin/auth"
	"mysql-sync-plugin/config"
	"mysql-sync-plugin/handler"
	"mysql-sync-plugin/logger"
	"mysql-sync-plugin/repository"
	"mysql-sync-plugin/service"
	"mysql-sync-plugin/static"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 获取MySQL连接字符串
	dsn := cfg.GetDSN()

	// 初始化日志存储
	if err := logger.GetStore().Init(dsn); err != nil {
		log.Fatalf("初始化日志数据库失败: %v", err)
	}
	defer logger.GetStore().Close()

	// 初始化认证存储
	if err := auth.GetStore().Init(dsn); err != nil {
		log.Fatalf("初始化认证数据库失败: %v", err)
	}
	defer auth.GetStore().Close()

	mainLog := logger.New("main")
	mainLog.Info("启动", "MySQL同步插件服务正在启动")

	// 初始化Repository
	authStore := auth.GetStore()
	repo := repository.NewSQLiteRepository(authStore.GetDB())

	// 初始化Service层
	cryptoService := service.NewCryptoService(cfg.SecretKey)
	userService := service.NewUserService(repo)
	datasourceService := service.NewDatasourceService(repo, cryptoService)
	permissionService := service.NewPermissionService(repo)

	// 设置Gin模式
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	r := gin.Default()
	r.RedirectTrailingSlash = false

	// CORS中间件(允许钉钉和飞书调用)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Ding-Docs-Timestamp, Ding-Docs-Signature, X-Base-Request-Timestamp, X-Base-Request-Nonce")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 创建处理器（使用新版本构造函数，支持数据源方案）
	h := handler.NewHandlerWithServices(datasourceService, repo)
	feishuH := handler.NewFeishuHandlerWithServices(datasourceService, repo)
	adminH := handler.NewAdminHandler()
	authH := handler.NewAuthHandler()
	userH := handler.NewUserHandler(userService)
	datasourceH := handler.NewDatasourceHandler(datasourceService)
	permissionH := handler.NewPermissionHandler(permissionService)
	userAuthH := handler.NewUserAuthHandler(userService, permissionService, authStore)

	// 通用的 index.html 处理函数
	serveIndexHTML := func(c *gin.Context, getData func() ([]byte, error)) {
		data, err := getData()
		if err != nil {
			c.Status(404)
			return
		}
		c.Data(200, "text/html; charset=utf-8", data)
	}

	// ==================== 公共接口 ====================

	// 健康检查
	r.GET("/health", h.Health)

	// 飞书 meta.json 配置文件
	r.GET("/meta.json", func(c *gin.Context) {
		data, err := static.GetFeishuMetaJSON()
		if err != nil {
			c.Status(404)
			return
		}
		c.Data(200, "application/json", data)
	})

	// ==================== 钉钉路由组 ====================

	// 钉钉前端静态文件
	r.GET("/dingtalk", func(c *gin.Context) { serveIndexHTML(c, static.GetDingtalkIndexHTML) })
	r.GET("/dingtalk/", func(c *gin.Context) { serveIndexHTML(c, static.GetDingtalkIndexHTML) })
	r.GET("/dingtalk/favicon.ico", func(c *gin.Context) {
		data, err := static.GetDingtalkFavicon()
		if err != nil {
			c.Status(404)
			return
		}
		c.Data(200, "image/x-icon", data)
	})
	r.StaticFS("/dingtalk/assets", static.GetDingtalkAssetsFS())

	dingtalkGroup := r.Group("/dingtalk")
	{

		// 钉钉API
		dingtalkAPI := dingtalkGroup.Group("/api")
		{
			// 用户认证API（无需登录）
			dingtalkAPI.POST("/auth/login", userAuthH.Login)
			dingtalkAPI.POST("/auth/logout", userAuthH.Logout)
			dingtalkAPI.GET("/auth/current", userAuthH.GetCurrentUser)

			// 用户数据源和表查询API（需要认证）
			dingtalkAPI.GET("/user/datasources", auth.UserAuthMiddleware(), userAuthH.GetUserDatasources)
			dingtalkAPI.GET("/user/datasources/:id/tables", auth.UserAuthMiddleware(), userAuthH.GetUserTables)

			// 前端配置页面使用的公共API
			dingtalkAPI.POST("/databases", h.GetDatabases)
			dingtalkAPI.POST("/tables", h.GetTables)
			dingtalkAPI.POST("/fields", h.GetFields)
			dingtalkAPI.POST("/preview_sql", h.PreviewSQL)
			// AI表格服务端调用的API（无需认证）
			dingtalkAPI.POST("/sheet_meta", h.SheetMeta)
			dingtalkAPI.POST("/records", h.Records)
		}
	}

	// ==================== 飞书路由组 ====================

	// 飞书前端静态文件
	r.GET("/feishu", func(c *gin.Context) { serveIndexHTML(c, static.GetFeishuIndexHTML) })
	r.GET("/feishu/", func(c *gin.Context) { serveIndexHTML(c, static.GetFeishuIndexHTML) })
	r.GET("/feishu/favicon.ico", func(c *gin.Context) {
		data, err := static.GetFeishuFavicon()
		if err != nil {
			c.Status(404)
			return
		}
		c.Data(200, "image/x-icon", data)
	})
	r.StaticFS("/feishu/assets", static.GetFeishuAssetsFS())

	feishuGroup := r.Group("/feishu")
	{

		// 飞书API
		feishuAPI := feishuGroup.Group("/api")
		{
			// 用户认证API（无需登录）
			feishuAPI.POST("/auth/login", userAuthH.Login)
			feishuAPI.POST("/auth/logout", userAuthH.Logout)
			feishuAPI.GET("/auth/current", userAuthH.GetCurrentUser)

			// 用户数据源和表查询API（需要认证）
			feishuAPI.GET("/user/datasources", auth.UserAuthMiddleware(), userAuthH.GetUserDatasources)
			feishuAPI.GET("/user/datasources/:id/tables", auth.UserAuthMiddleware(), userAuthH.GetUserTables)

			// 前端配置页面使用的公共API
			feishuAPI.POST("/databases", h.GetDatabases)
			feishuAPI.POST("/tables", h.GetTables)
			feishuAPI.POST("/fields", h.GetFields)
			feishuAPI.POST("/preview_sql", h.PreviewSQL)
			// 多维表格服务端调用的API（无需认证）
			feishuAPI.POST("/table_meta", feishuH.TableMeta)
			feishuAPI.POST("/records", feishuH.Records)
		}
	}

	// ==================== 管理后台路由组 ====================

	// 管理后台认证API（无需登录）
	adminAuth := r.Group("/admin/api")
	{
		adminAuth.POST("/login", authH.Login)
	}

	// 管理后台API路由组（需要登录）
	adminAPI := r.Group("/admin/api")
	adminAPI.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	})
	adminAPI.Use(auth.AdminAuthMiddleware())
	{
		// 认证相关
		adminAPI.POST("/logout", authH.Logout)
		adminAPI.GET("/user/current", authH.GetCurrentUser)
		adminAPI.POST("/user/password", authH.ChangePassword)

		// 日志管理
		adminAPI.GET("/logs", adminH.GetLogs)
		adminAPI.GET("/logs/stats", adminH.GetLogStats)
		adminAPI.POST("/logs/clean", adminH.CleanLogs)

		// 系统信息
		adminAPI.GET("/system/info", adminH.GetSystemInfo)

		// 用户管理（需要管理员权限）
		adminAPI.POST("/users", auth.RequireAdminRole(), userH.CreateUser)
		adminAPI.GET("/users", auth.RequireAdminRole(), userH.ListUsers)
		adminAPI.GET("/users/:id", auth.RequireAdminRole(), userH.GetUser)
		adminAPI.PUT("/users/:id", auth.RequireAdminRole(), userH.UpdateUser)
		adminAPI.DELETE("/users/:id", auth.RequireAdminRole(), userH.DeleteUser)
		adminAPI.PUT("/users/:id/status", auth.RequireAdminRole(), userH.UpdateUserStatus)
		adminAPI.PUT("/users/:id/password", auth.RequireAdminRole(), userH.ResetPassword)

		// 数据源管理（需要管理员权限）
		adminAPI.POST("/datasources", auth.RequireAdminRole(), datasourceH.CreateDatasource)
		adminAPI.GET("/datasources", auth.RequireAdminRole(), datasourceH.ListDatasources)
		adminAPI.GET("/datasources/:id", auth.RequireAdminRole(), datasourceH.GetDatasource)
		adminAPI.PUT("/datasources/:id", auth.RequireAdminRole(), datasourceH.UpdateDatasource)
		adminAPI.DELETE("/datasources/:id", auth.RequireAdminRole(), datasourceH.DeleteDatasource)
		adminAPI.POST("/datasources/:id/test", auth.RequireAdminRole(), datasourceH.TestConnection)

		// 数据源表管理（需要管理员权限）
		adminAPI.POST("/datasources/:id/tables", auth.RequireAdminRole(), datasourceH.CreateDatasourceTable)
		adminAPI.POST("/datasources/:id/tables/batch", auth.RequireAdminRole(), datasourceH.BatchCreateDatasourceTables)
		adminAPI.GET("/datasources/:id/tables", auth.RequireAdminRole(), datasourceH.ListDatasourceTables)
		adminAPI.GET("/datasource-tables/:id", auth.RequireAdminRole(), datasourceH.GetDatasourceTable)
		adminAPI.PUT("/datasource-tables/:id", auth.RequireAdminRole(), datasourceH.UpdateDatasourceTable)
		adminAPI.DELETE("/datasource-tables/:id", auth.RequireAdminRole(), datasourceH.DeleteDatasourceTable)
		adminAPI.GET("/datasource-tables/:id/fields", auth.RequireAdminRole(), datasourceH.GetFieldMappings)
		adminAPI.POST("/datasource-tables/:id/fields", auth.RequireAdminRole(), datasourceH.UpdateFieldMappings)

		// 数据源辅助接口（需要管理员权限）
		adminAPI.GET("/datasources/:id/databases", auth.RequireAdminRole(), datasourceH.GetDatabaseList)
		adminAPI.GET("/datasources/:id/db-tables", auth.RequireAdminRole(), datasourceH.GetTableList)
		adminAPI.GET("/datasources/:id/db-fields", auth.RequireAdminRole(), datasourceH.GetFieldList)
		adminAPI.POST("/datasources/:id/db-fields-from-sql", auth.RequireAdminRole(), datasourceH.GetFieldListFromSQL)

		// 权限管理（需要管理员权限）
		adminAPI.POST("/users/:id/datasources", auth.RequireAdminRole(), permissionH.GrantDatasourcePermissions)
		adminAPI.DELETE("/users/:id/datasources/:dsId", auth.RequireAdminRole(), permissionH.RevokeDatasourcePermission)
		adminAPI.DELETE("/users/:id/datasources", auth.RequireAdminRole(), permissionH.RevokeDatasourcePermissions)
		adminAPI.GET("/users/:id/datasources", auth.RequireAdminRole(), permissionH.ListUserDatasources)
		adminAPI.GET("/users/:id/datasources-with-permission", auth.RequireAdminRole(), permissionH.ListAllDatasourcesWithPermission)

		adminAPI.POST("/users/:id/tables", auth.RequireAdminRole(), permissionH.GrantTablePermissions)
		adminAPI.DELETE("/users/:id/tables/:tableId", auth.RequireAdminRole(), permissionH.RevokeTablePermission)
		adminAPI.DELETE("/users/:id/tables", auth.RequireAdminRole(), permissionH.RevokeTablePermissions)
		adminAPI.GET("/users/:id/tables", auth.RequireAdminRole(), permissionH.ListUserTables)
		adminAPI.GET("/users/:id/tables-with-permission", auth.RequireAdminRole(), permissionH.ListAllTablesWithPermission)
	}

	// ==================== 前端用户API ====================

	// 前端用户认证API（无需登录）
	userAuth := r.Group("/api/auth")
	{
		userAuth.POST("/login", userAuthH.Login)
	}

	// 前端用户API（需要登录）
	userAPI := r.Group("/api/user")
	userAPI.Use(auth.UserAuthMiddleware())
	{
		userAPI.POST("/logout", userAuthH.Logout)
		userAPI.GET("/current", userAuthH.GetCurrentUser)
		userAPI.GET("/datasources", userAuthH.GetUserDatasources)
		userAPI.GET("/datasources/:id/tables", userAuthH.GetUserTables)
	}

	// 管理后台静态文件服务
	r.StaticFS("/admin/assets", static.GetAdminAssetsFS())

	// 管理后台首页处理函数（需要禁用缓存）
	adminIndexHandler := func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		data, err := static.GetAdminIndexHTML()
		if err != nil {
			c.Status(404)
			return
		}
		c.Data(200, "text/html; charset=utf-8", data)
	}
	r.GET("/admin", adminIndexHandler)
	r.GET("/admin/", adminIndexHandler)

	// SPA 路由支持：管理后台前端路由回退
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 管理后台 SPA 路由
		if len(path) >= 6 && path[:6] == "/admin" {
			// 排除静态资源和 API 路径
			if len(path) > 13 && path[:13] == "/admin/assets" {
				c.Status(404)
				return
			}
			if len(path) > 10 && path[:10] == "/admin/api" {
				c.Status(404)
				return
			}
			adminIndexHandler(c)
			return
		}

		c.Status(404)
	})

	// 启动服务
	addr := ":" + cfg.ServerPort
	mainLog.Infof("启动", "服务启动在端口 %s", cfg.ServerPort)
	mainLog.Info("启动", "========== 路由信息 ==========")
	mainLog.Info("启动", "钉钉前端: http://localhost:"+cfg.ServerPort+"/dingtalk/")
	mainLog.Info("启动", "飞书前端: http://localhost:"+cfg.ServerPort+"/feishu/")
	mainLog.Info("启动", "管理后台: http://localhost:"+cfg.ServerPort+"/admin")
	mainLog.Info("启动", "钉钉API: /dingtalk/api/*")
	mainLog.Info("启动", "飞书API: /feishu/api/*")
	mainLog.Info("启动", "===============================")

	if err := r.Run(addr); err != nil {
		mainLog.Errorf("启动", "服务启动失败: %v", err)
		log.Fatalf("服务启动失败: %v", err)
	}
}

