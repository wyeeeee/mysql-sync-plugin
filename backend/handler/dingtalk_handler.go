package handler

import (
	"encoding/json"
	"fmt"
	"mysql-sync-plugin/logger"
	"mysql-sync-plugin/models"
	"mysql-sync-plugin/repository"
	"mysql-sync-plugin/service"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler API处理器
type Handler struct {
	mysqlService      *service.MySQLService
	datasourceService *service.DatasourceService
	repo              repository.Repository
	log               *logger.Logger
}

// NewHandler 创建处理器实例（老版本，兼容旧代码）
func NewHandler() *Handler {
	return &Handler{
		mysqlService: service.NewMySQLService(),
		log:          logger.New("dingtalk-api"),
	}
}

// NewHandlerWithServices 创建处理器实例（新版本，支持数据源方案）
func NewHandlerWithServices(datasourceService *service.DatasourceService, repo repository.Repository) *Handler {
	return &Handler{
		mysqlService:      service.NewMySQLService(),
		datasourceService: datasourceService,
		repo:              repo,
		log:               logger.New("dingtalk-api"),
	}
}

// resolveConfig 解析配置：支持新格式（tableId）和老格式（完整配置）
func (h *Handler) resolveConfig(config *models.MySQLConfig) (*models.MySQLConfig, error) {
	// 新格式：通过 tableId 查询配置
	if config.IsNewFormat() {
		if h.datasourceService == nil || h.repo == nil {
			return nil, fmt.Errorf("服务未初始化，无法使用 tableId 模式")
		}

		// 获取表配置
		table, err := h.datasourceService.GetDatasourceTableByID(config.TableID)
		if err != nil {
			return nil, fmt.Errorf("获取表配置失败: %w", err)
		}

		// 获取数据源配置（包含解密后的密码）
		ds, err := h.datasourceService.GetDatasourceByIDWithPassword(table.DatasourceID)
		if err != nil {
			return nil, fmt.Errorf("获取数据源配置失败: %w", err)
		}

		// 获取字段映射
		fieldMappings, err := h.repo.ListFieldMappings(table.ID)
		if err != nil {
			return nil, fmt.Errorf("获取字段映射失败: %w", err)
		}

		// 转换字段映射格式
		var mappings []models.FieldMapping
		for _, fm := range fieldMappings {
			mappings = append(mappings, models.FieldMapping{
				MysqlField: fm.FieldName,
				AliasField: fm.FieldAlias,
			})
		}

		// 构建完整的 MySQLConfig
		resolvedConfig := &models.MySQLConfig{
			Host:          ds.Host,
			Port:          ds.Port,
			Database:      ds.DatabaseName,
			Username:      ds.Username,
			Password:      ds.Password,
			Table:         table.TableName,
			QueryMode:     table.QueryMode,
			CustomSQL:     table.CustomSQL,
			FieldMappings: mappings,
		}

		return resolvedConfig, nil
	}

	// 老格式：直接使用传入的配置
	return config, nil
}

// SheetMeta 获取表结构
func (h *Handler) SheetMeta(c *gin.Context) {
	start := time.Now()
	ip := c.ClientIP()

	var req models.SheetMetaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取表结构", "参数解析失败", err.Error(), ip, c.GetHeader("User-Agent"), 0)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "请求参数错误: " + err.Error(),
		})
		return
	}

	// 解析配置
	var config models.MySQLConfig
	if err := json.Unmarshal([]byte(req.Params), &config); err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取表结构", "配置解析失败", err.Error(), ip, c.GetHeader("User-Agent"), 0)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "配置解析失败: " + err.Error(),
		})
		return
	}

	// 解析配置（支持新格式和老格式）
	resolvedConfig, err := h.resolveConfig(&config)
	if err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取表结构", "解析配置失败: "+err.Error(), "", ip, c.GetHeader("User-Agent"), 0)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeConfigError,
			Msg:  "配置解析失败: " + err.Error(),
		})
		return
	}

	// 重新序列化配置
	resolvedParams, _ := json.Marshal(resolvedConfig)

	// 构建实际执行的SQL
	var executedSQL string
	if resolvedConfig.QueryMode == "sql" && resolvedConfig.CustomSQL != "" {
		executedSQL = fmt.Sprintf("SELECT * FROM (%s) AS t LIMIT 1", resolvedConfig.CustomSQL)
	} else {
		executedSQL = fmt.Sprintf("SELECT COLUMN_NAME, DATA_TYPE, COLUMN_KEY, COLUMN_COMMENT FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = '%s'", resolvedConfig.Database, resolvedConfig.Table)
	}
	detail := fmt.Sprintf("主机: %s:%d, 数据库: %s, 表: %s, 模式: %s\nSQL: %s", resolvedConfig.Host, resolvedConfig.Port, resolvedConfig.Database, resolvedConfig.Table, resolvedConfig.QueryMode, executedSQL)

	// 构建新的请求
	newReq := &models.SheetMetaRequest{
		RequestID: req.RequestID,
		Params:    string(resolvedParams),
		Context:   req.Context,
	}

	// 调用服务层
	data, err := h.mysqlService.GetSheetMeta(newReq)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取表结构", "获取表结构失败: "+err.Error(), detail, ip, c.GetHeader("User-Agent"), duration)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取表结构失败: " + err.Error(),
		})
		return
	}

	h.log.LogWithRequest(logger.LevelInfo, "获取表结构", fmt.Sprintf("成功获取 %d 个字段", len(data.Fields)), detail, ip, c.GetHeader("User-Agent"), duration)

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: data,
	})
}

// Records 获取表记录
func (h *Handler) Records(c *gin.Context) {
	start := time.Now()
	ip := c.ClientIP()

	var req models.RecordsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取记录", "参数解析失败", err.Error(), ip, c.GetHeader("User-Agent"), 0)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "请求参数错误: " + err.Error(),
		})
		return
	}

	// 解析配置
	var config models.MySQLConfig
	if err := json.Unmarshal([]byte(req.Params), &config); err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取记录", "配置解析失败", err.Error(), ip, c.GetHeader("User-Agent"), 0)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "配置解析失败: " + err.Error(),
		})
		return
	}

	// 解析配置（支持新格式和老格式）
	resolvedConfig, err := h.resolveConfig(&config)
	if err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取记录", "解析配置失败: "+err.Error(), "", ip, c.GetHeader("User-Agent"), 0)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeConfigError,
			Msg:  "配置解析失败: " + err.Error(),
		})
		return
	}

	// 重新序列化配置
	resolvedParams, _ := json.Marshal(resolvedConfig)

	// 解析分页参数用于构建SQL
	offset := 0
	if req.NextToken != "" {
		parts := strings.Split(req.NextToken, ":")
		if len(parts) == 2 {
			offset, _ = strconv.Atoi(parts[1])
		}
	}
	maxResults := req.MaxResults
	if maxResults <= 0 {
		maxResults = 300
	}

	// 构建实际执行的SQL
	var executedSQL string
	if resolvedConfig.QueryMode == "sql" && resolvedConfig.CustomSQL != "" {
		executedSQL = fmt.Sprintf("SELECT * FROM (%s) AS t LIMIT %d OFFSET %d", resolvedConfig.CustomSQL, maxResults, offset)
	} else {
		executedSQL = fmt.Sprintf("SELECT * FROM `%s` LIMIT %d OFFSET %d", resolvedConfig.Table, maxResults, offset)
	}
	detail := fmt.Sprintf("主机: %s:%d, 数据库: %s, 表: %s, 模式: %s\nSQL: %s",
		resolvedConfig.Host, resolvedConfig.Port, resolvedConfig.Database, resolvedConfig.Table, resolvedConfig.QueryMode, executedSQL)

	// 构建新的请求
	newReq := &models.RecordsRequest{
		RequestID:  req.RequestID,
		MaxResults: req.MaxResults,
		NextToken:  req.NextToken,
		Params:     string(resolvedParams),
		Context:    req.Context,
	}

	// 调用服务层
	data, err := h.mysqlService.GetRecords(newReq)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取记录", "获取表记录失败: "+err.Error(), detail, ip, c.GetHeader("User-Agent"), duration)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取表记录失败: " + err.Error(),
		})
		return
	}

	h.log.LogWithRequest(logger.LevelInfo, "获取记录",
		fmt.Sprintf("成功获取 %d 条记录, 总数: %d, 还有更多: %v", len(data.Records), data.Total, data.HasMore),
		detail, ip, c.GetHeader("User-Agent"), duration)

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: data,
	})
}

// GetDatabases 获取数据库列表(前端辅助接口)
func (h *Handler) GetDatabases(c *gin.Context) {
	start := time.Now()
	ip := c.ClientIP()

	var config models.MySQLConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取数据库列表", "参数解析失败", err.Error(), ip, c.GetHeader("User-Agent"), 0)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "请求参数错误: " + err.Error(),
		})
		return
	}

	detail := fmt.Sprintf("主机: %s:%d", config.Host, config.Port)
	h.log.InfoWithDetail("获取数据库列表", "尝试连接MySQL服务器", detail)

	databases, err := h.mysqlService.GetDatabases(&config)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取数据库列表", "连接失败: "+err.Error(), detail, ip, c.GetHeader("User-Agent"), duration)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取数据库列表失败: " + err.Error(),
		})
		return
	}

	h.log.LogWithRequest(logger.LevelInfo, "获取数据库列表", fmt.Sprintf("成功获取 %d 个数据库", len(databases)), detail, ip, c.GetHeader("User-Agent"), duration)

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: databases,
	})
}

// GetTables 获取数据表列表(前端辅助接口)
func (h *Handler) GetTables(c *gin.Context) {
	start := time.Now()
	ip := c.ClientIP()

	var config models.MySQLConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取表列表", "参数解析失败", err.Error(), ip, c.GetHeader("User-Agent"), 0)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "请求参数错误: " + err.Error(),
		})
		return
	}

	detail := fmt.Sprintf("主机: %s:%d, 数据库: %s", config.Host, config.Port, config.Database)

	tables, err := h.mysqlService.GetTables(&config)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取表列表", "获取失败: "+err.Error(), detail, ip, c.GetHeader("User-Agent"), duration)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取数据表列表失败: " + err.Error(),
		})
		return
	}

	h.log.LogWithRequest(logger.LevelInfo, "获取表列表", fmt.Sprintf("成功获取 %d 个表", len(tables)), detail, ip, c.GetHeader("User-Agent"), duration)

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: tables,
	})
}

// GetFields 获取表字段(前端辅助接口)
func (h *Handler) GetFields(c *gin.Context) {
	start := time.Now()
	ip := c.ClientIP()

	var config models.MySQLConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取字段", "参数解析失败", err.Error(), ip, c.GetHeader("User-Agent"), 0)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "请求参数错误: " + err.Error(),
		})
		return
	}

	detail := fmt.Sprintf("主机: %s:%d, 数据库: %s, 表: %s", config.Host, config.Port, config.Database, config.Table)

	fields, err := h.mysqlService.GetFields(&config)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取字段", "获取失败: "+err.Error(), detail, ip, c.GetHeader("User-Agent"), duration)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取字段列表失败: " + err.Error(),
		})
		return
	}

	h.log.LogWithRequest(logger.LevelInfo, "获取字段", fmt.Sprintf("成功获取 %d 个字段", len(fields)), detail, ip, c.GetHeader("User-Agent"), duration)

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: fields,
	})
}

// PreviewSQL 预览SQL执行结果（获取字段列表）
func (h *Handler) PreviewSQL(c *gin.Context) {
	start := time.Now()
	ip := c.ClientIP()

	var config models.MySQLConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		h.log.LogWithRequest(logger.LevelError, "预览SQL", "参数解析失败", err.Error(), ip, c.GetHeader("User-Agent"), 0)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "请求参数错误: " + err.Error(),
		})
		return
	}

	if config.CustomSQL == "" {
		h.log.LogWithRequest(logger.LevelWarn, "预览SQL", "SQL语句为空", "", ip, c.GetHeader("User-Agent"), 0)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "SQL语句不能为空",
		})
		return
	}

	detail := fmt.Sprintf("主机: %s:%d, 数据库: %s, SQL: %s", config.Host, config.Port, config.Database, config.CustomSQL)
	h.log.InfoWithDetail("预览SQL", "开始执行SQL预览", detail)

	fields, err := h.mysqlService.PreviewSQL(&config)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		h.log.LogWithRequest(logger.LevelError, "预览SQL", "SQL执行失败: "+err.Error(), detail, ip, c.GetHeader("User-Agent"), duration)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "SQL执行失败: " + err.Error(),
		})
		return
	}

	h.log.LogWithRequest(logger.LevelInfo, "预览SQL", fmt.Sprintf("SQL预览成功, 返回 %d 个字段", len(fields)), detail, ip, c.GetHeader("User-Agent"), duration)

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: fields,
	})
}

// Health 健康检查
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "mysql-sync-plugin",
	})
}
