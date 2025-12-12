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

// FeishuHandler 飞书多维表格API处理器
type FeishuHandler struct {
	mysqlService      *service.MySQLService
	datasourceService *service.DatasourceService
	repo              repository.Repository
	log               *logger.Logger
}

// NewFeishuHandler 创建飞书处理器实例（老版本，兼容旧代码）
func NewFeishuHandler() *FeishuHandler {
	return &FeishuHandler{
		mysqlService: service.NewMySQLService(),
		log:          logger.New("feishu-api"),
	}
}

// NewFeishuHandlerWithServices 创建飞书处理器实例（新版本，支持数据源方案）
func NewFeishuHandlerWithServices(datasourceService *service.DatasourceService, repo repository.Repository) *FeishuHandler {
	return &FeishuHandler{
		mysqlService:      service.NewMySQLService(),
		datasourceService: datasourceService,
		repo:              repo,
		log:               logger.New("feishu-api"),
	}
}

// resolveConfig 解析配置：支持新格式（tableId）和老格式（完整配置）
func (h *FeishuHandler) resolveConfig(config *models.MySQLConfig) (*models.MySQLConfig, []models.FieldMapping, error) {
	// 新格式：通过 tableId 查询配置
	if config.IsNewFormat() {
		if h.datasourceService == nil || h.repo == nil {
			return nil, nil, fmt.Errorf("服务未初始化，无法使用 tableId 模式")
		}

		// 获取表配置
		table, err := h.datasourceService.GetDatasourceTableByID(config.TableID)
		if err != nil {
			return nil, nil, fmt.Errorf("获取表配置失败: %w", err)
		}

		// 获取数据源配置（包含解密后的密码）
		ds, err := h.datasourceService.GetDatasourceByIDWithPassword(table.DatasourceID)
		if err != nil {
			return nil, nil, fmt.Errorf("获取数据源配置失败: %w", err)
		}

		// 获取字段映射
		fieldMappings, err := h.repo.ListFieldMappings(table.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("获取字段映射失败: %w", err)
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

		return resolvedConfig, mappings, nil
	}

	// 老格式：直接使用传入的配置
	return config, config.FieldMappings, nil
}

// TableMeta 获取表结构（飞书格式）
func (h *FeishuHandler) TableMeta(c *gin.Context) {
	start := time.Now()
	ip := c.ClientIP()

	var req models.FeishuTableMetaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取表结构", "参数解析失败", err.Error(), ip, c.GetHeader("User-Agent"), 0)
		c.JSON(http.StatusOK, models.FeishuResponse{
			Code: models.FeishuCodeConfigError,
			Msg:  models.NewFeishuErrorMsg("请求参数错误: "+err.Error(), "Invalid request parameters: "+err.Error()),
		})
		return
	}

	// 解析params中的datasourceConfig
	var feishuParams models.FeishuParams
	if err := json.Unmarshal([]byte(req.Params), &feishuParams); err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取表结构", "params解析失败", err.Error(), ip, c.GetHeader("User-Agent"), 0)
		c.JSON(http.StatusOK, models.FeishuResponse{
			Code: models.FeishuCodeConfigError,
			Msg:  models.NewFeishuErrorMsg("参数格式错误", "Invalid params format"),
		})
		return
	}

	// 解析context（飞书的context是JSON字符串）
	var feishuContext models.FeishuRequestContext
	if req.Context != "" {
		if err := json.Unmarshal([]byte(req.Context), &feishuContext); err != nil {
			h.log.LogWithRequest(logger.LevelWarn, "获取表结构", "context解析失败，使用空上下文", err.Error(), ip, c.GetHeader("User-Agent"), 0)
		}
	}

	// 解析MySQL配置
	var config models.MySQLConfig
	if err := json.Unmarshal([]byte(feishuParams.DatasourceConfig), &config); err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取表结构", "数据源配置解析失败", err.Error(), ip, c.GetHeader("User-Agent"), 0)
		c.JSON(http.StatusOK, models.FeishuResponse{
			Code: models.FeishuCodeConfigError,
			Msg:  models.NewFeishuErrorMsg("数据源配置格式错误", "Invalid datasource config format"),
		})
		return
	}

	// 解析配置（支持新格式和老格式）
	resolvedConfig, fieldMappings, err := h.resolveConfig(&config)
	if err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取表结构", "解析配置失败: "+err.Error(), "", ip, c.GetHeader("User-Agent"), 0)
		c.JSON(http.StatusOK, models.FeishuResponse{
			Code: models.FeishuCodeConfigError,
			Msg:  models.NewFeishuErrorMsg("配置解析失败: "+err.Error(), "Failed to resolve config: "+err.Error()),
		})
		return
	}

	// 飞书要求 fieldID 只能包含英文数字下划线
	// 因此不能让服务层应用别名到 ID，需要清空 fieldMappings
	resolvedConfig.FieldMappings = nil
	configWithoutMappings, _ := json.Marshal(resolvedConfig)

	// 构建日志详情
	var executedSQL string
	if resolvedConfig.QueryMode == "sql" && resolvedConfig.CustomSQL != "" {
		executedSQL = fmt.Sprintf("SELECT * FROM (%s) AS t LIMIT 1", resolvedConfig.CustomSQL)
	} else {
		executedSQL = fmt.Sprintf("SELECT COLUMN_NAME, DATA_TYPE, COLUMN_KEY, COLUMN_COMMENT FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = '%s'", resolvedConfig.Database, resolvedConfig.Table)
	}
	detail := fmt.Sprintf("主机: %s:%d, 数据库: %s, 表: %s, 模式: %s\nSQL: %s", resolvedConfig.Host, resolvedConfig.Port, resolvedConfig.Database, resolvedConfig.Table, resolvedConfig.QueryMode, executedSQL)

	h.log.InfoWithDetail("获取表结构", "开始获取表结构", detail)

	// 构建钉钉格式的请求，复用服务层
	dingtalkReq := &models.SheetMetaRequest{
		RequestID: feishuContext.Bitable.LogID,
		Params:    string(configWithoutMappings),
		Context: models.Context{
			UnionID: feishuContext.ScriptArgs.BaseOpenID,
			CorpID:  feishuContext.TenantKey,
		},
	}

	// 调用服务层
	data, err := h.mysqlService.GetSheetMeta(dingtalkReq)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取表结构", "获取表结构失败: "+err.Error(), detail, ip, c.GetHeader("User-Agent"), duration)
		c.JSON(http.StatusOK, models.FeishuResponse{
			Code: models.FeishuCodeSystemError,
			Msg:  models.NewFeishuErrorMsg("获取表结构失败: "+err.Error(), "Failed to get table meta: "+err.Error()),
		})
		return
	}

	// 转换为飞书格式，并应用字段映射到 fieldName
	feishuData := models.ConvertToFeishuTableMetaWithMappings(data, fieldMappings)

	h.log.LogWithRequest(logger.LevelInfo, "获取表结构", fmt.Sprintf("成功获取 %d 个字段", len(feishuData.Fields)), detail, ip, c.GetHeader("User-Agent"), duration)

	c.JSON(http.StatusOK, models.FeishuResponse{
		Code: models.FeishuCodeSuccess,
		Data: feishuData,
	})
}

// Records 获取表记录（飞书格式）
func (h *FeishuHandler) Records(c *gin.Context) {
	start := time.Now()
	ip := c.ClientIP()

	var req models.FeishuRecordsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取记录", "参数解析失败", err.Error(), ip, c.GetHeader("User-Agent"), 0)
		c.JSON(http.StatusOK, models.FeishuResponse{
			Code: models.FeishuCodeConfigError,
			Msg:  models.NewFeishuErrorMsg("请求参数错误: "+err.Error(), "Invalid request parameters: "+err.Error()),
		})
		return
	}

	// 解析params
	var feishuParams models.FeishuParams
	if err := json.Unmarshal([]byte(req.Params), &feishuParams); err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取记录", "params解析失败", err.Error(), ip, c.GetHeader("User-Agent"), 0)
		c.JSON(http.StatusOK, models.FeishuResponse{
			Code: models.FeishuCodeConfigError,
			Msg:  models.NewFeishuErrorMsg("参数格式错误", "Invalid params format"),
		})
		return
	}

	// 解析context（飞书的context是JSON字符串）
	var feishuContext models.FeishuRequestContext
	if req.Context != "" {
		if err := json.Unmarshal([]byte(req.Context), &feishuContext); err != nil {
			h.log.LogWithRequest(logger.LevelWarn, "获取记录", "context解析失败，使用空上下文", err.Error(), ip, c.GetHeader("User-Agent"), 0)
		}
	}

	// 解析MySQL配置
	var config models.MySQLConfig
	if err := json.Unmarshal([]byte(feishuParams.DatasourceConfig), &config); err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取记录", "数据源配置解析失败", err.Error(), ip, c.GetHeader("User-Agent"), 0)
		c.JSON(http.StatusOK, models.FeishuResponse{
			Code: models.FeishuCodeConfigError,
			Msg:  models.NewFeishuErrorMsg("数据源配置格式错误", "Invalid datasource config format"),
		})
		return
	}

	// 解析配置（支持新格式和老格式）
	resolvedConfig, _, err := h.resolveConfig(&config)
	if err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取记录", "解析配置失败: "+err.Error(), "", ip, c.GetHeader("User-Agent"), 0)
		c.JSON(http.StatusOK, models.FeishuResponse{
			Code: models.FeishuCodeConfigError,
			Msg:  models.NewFeishuErrorMsg("配置解析失败: "+err.Error(), "Failed to resolve config: "+err.Error()),
		})
		return
	}

	// 飞书要求 fieldID 只能包含英文数字下划线
	// 因此不能让服务层应用别名到字段key，需要清空 fieldMappings
	resolvedConfig.FieldMappings = nil
	configWithoutMappings, _ := json.Marshal(resolvedConfig)

	// 转换分页参数：飞书pageToken -> 钉钉nextToken
	nextToken := ""
	if feishuParams.PageToken != "" {
		nextToken = feishuParams.PageToken
	}

	// 转换分页大小
	maxResults := feishuParams.MaxPageSize
	if maxResults <= 0 {
		maxResults = 300
	}

	// 解析分页参数用于构建SQL
	offset := 0
	if nextToken != "" {
		parts := strings.Split(nextToken, ":")
		if len(parts) == 2 {
			offset, _ = strconv.Atoi(parts[1])
		}
	}

	// 构建日志详情
	var executedSQL string
	if resolvedConfig.QueryMode == "sql" && resolvedConfig.CustomSQL != "" {
		executedSQL = fmt.Sprintf("SELECT * FROM (%s) AS t LIMIT %d OFFSET %d", resolvedConfig.CustomSQL, maxResults, offset)
	} else {
		executedSQL = fmt.Sprintf("SELECT * FROM `%s` LIMIT %d OFFSET %d", resolvedConfig.Table, maxResults, offset)
	}
	detail := fmt.Sprintf("主机: %s:%d, 数据库: %s, 表: %s, 模式: %s\nSQL: %s",
		resolvedConfig.Host, resolvedConfig.Port, resolvedConfig.Database, resolvedConfig.Table, resolvedConfig.QueryMode, executedSQL)

	h.log.InfoWithDetail("获取记录", "开始获取表记录", detail)

	// 构建钉钉格式的请求，复用服务层
	dingtalkReq := &models.RecordsRequest{
		RequestID:  feishuContext.Bitable.LogID,
		MaxResults: maxResults,
		NextToken:  nextToken,
		Params:     string(configWithoutMappings),
		Context: models.Context{
			UnionID: feishuContext.ScriptArgs.BaseOpenID,
			CorpID:  feishuContext.TenantKey,
		},
	}

	// 调用服务层获取记录
	data, err := h.mysqlService.GetRecords(dingtalkReq)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取记录", "获取表记录失败: "+err.Error(), detail, ip, c.GetHeader("User-Agent"), duration)
		c.JSON(http.StatusOK, models.FeishuResponse{
			Code: models.FeishuCodeSystemError,
			Msg:  models.NewFeishuErrorMsg("获取表记录失败: "+err.Error(), "Failed to get records: "+err.Error()),
		})
		return
	}

	// 获取表结构用于字段映射
	metaReq := &models.SheetMetaRequest{
		Params: string(configWithoutMappings),
	}
	metaData, err := h.mysqlService.GetSheetMeta(metaReq)
	if err != nil {
		h.log.LogWithRequest(logger.LevelError, "获取记录", "获取表结构失败: "+err.Error(), detail, ip, c.GetHeader("User-Agent"), duration)
		c.JSON(http.StatusOK, models.FeishuResponse{
			Code: models.FeishuCodeSystemError,
			Msg:  models.NewFeishuErrorMsg("获取表结构失败: "+err.Error(), "Failed to get table meta: "+err.Error()),
		})
		return
	}

	// 转换为飞书格式，传入字段列表以保证顺序一致
	feishuData := models.ConvertToFeishuRecords(data, metaData.Fields)

	h.log.LogWithRequest(logger.LevelInfo, "获取记录",
		fmt.Sprintf("成功获取 %d 条记录, 还有更多: %v", len(feishuData.Records), feishuData.HasMore),
		detail, ip, c.GetHeader("User-Agent"), duration)

	c.JSON(http.StatusOK, models.FeishuResponse{
		Code: models.FeishuCodeSuccess,
		Data: feishuData,
	})
}
