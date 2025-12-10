package handler

import (
	"encoding/json"
	"fmt"
	"mysql-sync-plugin/logger"
	"mysql-sync-plugin/models"
	"mysql-sync-plugin/service"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// FeishuHandler 飞书多维表格API处理器
type FeishuHandler struct {
	mysqlService *service.MySQLService
	log          *logger.Logger
}

// NewFeishuHandler 创建飞书处理器实例
func NewFeishuHandler() *FeishuHandler {
	return &FeishuHandler{
		mysqlService: service.NewMySQLService(),
		log:          logger.New("feishu-api"),
	}
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

	// 保存字段映射，稍后用于飞书格式转换
	fieldMappings := config.FieldMappings

	// 飞书要求 fieldID 只能包含英文数字下划线
	// 因此不能让服务层应用别名到 ID，需要清空 fieldMappings
	config.FieldMappings = nil
	configWithoutMappings, _ := json.Marshal(config)

	// 构建日志详情
	var executedSQL string
	if config.QueryMode == "sql" && config.CustomSQL != "" {
		executedSQL = fmt.Sprintf("SELECT * FROM (%s) AS t LIMIT 1", config.CustomSQL)
	} else {
		executedSQL = fmt.Sprintf("SELECT COLUMN_NAME, DATA_TYPE, COLUMN_KEY, COLUMN_COMMENT FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = '%s'", config.Database, config.Table)
	}
	detail := fmt.Sprintf("[飞书] 主机: %s:%d, 数据库: %s, 表: %s, 模式: %s\nSQL: %s", config.Host, config.Port, config.Database, config.Table, config.QueryMode, executedSQL)

	h.log.InfoWithDetail("获取表结构", "开始获取表结构(飞书)", detail)

	// 构建钉钉格式的请求，复用服务层
	// 使用不含字段映射的配置，让 fieldID 保持原始英文字段名
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

	h.log.LogWithRequest(logger.LevelInfo, "获取表结构", fmt.Sprintf("[飞书] 成功获取 %d 个字段", len(feishuData.Fields)), detail, ip, c.GetHeader("User-Agent"), duration)

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

	// 飞书要求 fieldID 只能包含英文数字下划线
	// 因此不能让服务层应用别名到字段key，需要清空 fieldMappings
	config.FieldMappings = nil
	configWithoutMappings, _ := json.Marshal(config)

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
	if config.QueryMode == "sql" && config.CustomSQL != "" {
		executedSQL = fmt.Sprintf("SELECT * FROM (%s) AS t LIMIT %d OFFSET %d", config.CustomSQL, maxResults, offset)
	} else {
		executedSQL = fmt.Sprintf("SELECT * FROM `%s` LIMIT %d OFFSET %d", config.Table, maxResults, offset)
	}
	detail := fmt.Sprintf("[飞书] 主机: %s:%d, 数据库: %s, 表: %s, 模式: %s\nSQL: %s",
		config.Host, config.Port, config.Database, config.Table, config.QueryMode, executedSQL)

	h.log.InfoWithDetail("获取记录", "开始获取表记录(飞书)", detail)

	// 构建钉钉格式的请求，复用服务层
	// 使用不含字段映射的配置，让字段key保持原始英文字段名
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

	// 调用服务层
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

	// 转换为飞书格式
	feishuData := models.ConvertToFeishuRecords(data)

	h.log.LogWithRequest(logger.LevelInfo, "获取记录",
		fmt.Sprintf("[飞书] 成功获取 %d 条记录, 还有更多: %v", len(feishuData.Records), feishuData.HasMore),
		detail, ip, c.GetHeader("User-Agent"), duration)

	c.JSON(http.StatusOK, models.FeishuResponse{
		Code: models.FeishuCodeSuccess,
		Data: feishuData,
	})
}
