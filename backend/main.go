package main

import (
	"io/fs"
	"log"
	"mysql-sync-plugin/auth"
	"mysql-sync-plugin/config"
	"mysql-sync-plugin/handler"
	"mysql-sync-plugin/logger"
	"mysql-sync-plugin/repository"
	"mysql-sync-plugin/service"
	"mysql-sync-plugin/static"
	"net/http"
	"path/filepath"
	"strings"

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

	// 获取嵌入的静态文件系统
	adminEmbedFS := static.GetAdminEmbedFS()
	dingtalkEmbedFS := static.GetDingtalkEmbedFS()
	feishuEmbedFS := static.GetFeishuEmbedFS()

	// 创建子文件系统用于静态资源服务
	adminSubFS, _ := fs.Sub(adminEmbedFS, "admin")
	dingtalkSubFS, _ := fs.Sub(dingtalkEmbedFS, "dingtalk")
	feishuSubFS, _ := fs.Sub(feishuEmbedFS, "feishu")

	// 根据文件扩展名获取 Content-Type
	getContentType := func(filePath string) string {
		ext := strings.ToLower(filepath.Ext(filePath))
		switch ext {
		case ".html":
			return "text/html; charset=utf-8"
		case ".css":
			return "text/css; charset=utf-8"
		case ".js":
			return "application/javascript; charset=utf-8"
		case ".json":
			return "application/json; charset=utf-8"
		case ".png":
			return "image/png"
		case ".jpg", ".jpeg":
			return "image/jpeg"
		case ".gif":
			return "image/gif"
		case ".svg":
			return "image/svg+xml"
		case ".ico":
			return "image/x-icon"
		case ".woff":
			return "font/woff"
		case ".woff2":
			return "font/woff2"
		case ".ttf":
			return "font/ttf"
		case ".eot":
			return "application/vnd.ms-fontobject"
		default:
			return "application/octet-stream"
		}
	}

	// 辅助函数：从嵌入文件系统读取文件并返回
	serveEmbedFile := func(c *gin.Context, embedFS fs.FS, filePath string, contentType string) {
		data, err := fs.ReadFile(embedFS, filePath)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, contentType, data)
	}

	// 辅助函数：从嵌入文件系统读取文件并自动判断 Content-Type
	serveEmbedFileAuto := func(c *gin.Context, embedFS fs.FS, filePath string) {
		data, err := fs.ReadFile(embedFS, filePath)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, getContentType(filePath), data)
	}

	// ==================== 公共接口 ====================

	// 健康检查
	r.GET("/health", h.Health)

	// 飞书 meta.json 配置文件（从嵌入文件系统读取）
	r.GET("/meta.json", func(c *gin.Context) {
		serveEmbedFile(c, feishuSubFS, "meta.json", "application/json")
	})

	// ==================== 钉钉路由组 ====================

	// 钉钉首页处理函数
	dingtalkIndexHandler := func(c *gin.Context) {
		serveEmbedFile(c, dingtalkSubFS, "index.html", "text/html; charset=utf-8")
	}
	r.GET("/dingtalk", dingtalkIndexHandler)

	dingtalkGroup := r.Group("/dingtalk")
	{
		// 钉钉前端静态文件（从嵌入文件系统）
		dingtalkGroup.GET("/assets/*filepath", func(c *gin.Context) {
			fp := c.Param("filepath")
			serveEmbedFileAuto(c, dingtalkSubFS, "assets"+fp)
		})
		dingtalkGroup.GET("/", dingtalkIndexHandler)
		dingtalkGroup.GET("/favicon.ico", func(c *gin.Context) {
			serveEmbedFileAuto(c, dingtalkSubFS, "favicon.ico")
		})

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

	// 飞书首页处理函数
	feishuIndexHandler := func(c *gin.Context) {
		serveEmbedFile(c, feishuSubFS, "index.html", "text/html; charset=utf-8")
	}
	r.GET("/feishu", feishuIndexHandler)

	feishuGroup := r.Group("/feishu")
	{
		// 飞书前端静态文件（从嵌入文件系统）
		feishuGroup.GET("/assets/*filepath", func(c *gin.Context) {
			fp := c.Param("filepath")
			serveEmbedFileAuto(c, feishuSubFS, "assets"+fp)
		})
		feishuGroup.GET("/", feishuIndexHandler)
		feishuGroup.GET("/favicon.ico", func(c *gin.Context) {
			serveEmbedFileAuto(c, feishuSubFS, "favicon.ico")
		})

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

	// 管理后台静态文件服务（从嵌入文件系统）
	r.GET("/admin/assets/*filepath", func(c *gin.Context) {
		fp := c.Param("filepath")
		c.Header("Cache-Control", "no-cache, must-revalidate")
		serveEmbedFileAuto(c, adminSubFS, "assets"+fp)
	})

	// 管理后台首页处理函数
	adminIndexHandler := func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		serveEmbedFile(c, adminSubFS, "index.html", "text/html; charset=utf-8")
	}
	r.GET("/admin", adminIndexHandler)
	r.GET("/admin/", adminIndexHandler)

	// SPA 路由支持：处理前端路由路径
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 管理后台 SPA 路由
		if len(path) >= 6 && path[:6] == "/admin" {
			if len(path) > 13 && path[:13] == "/admin/assets" {
				c.Status(404)
				return
			}
			if len(path) > 10 && path[:10] == "/admin/api" {
				c.Status(404)
				return
			}
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
			serveEmbedFile(c, adminSubFS, "index.html", "text/html; charset=utf-8")
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

