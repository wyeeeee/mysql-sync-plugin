package handler

import (
	"mysql-sync-plugin/models"
	"mysql-sync-plugin/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler API处理器
type Handler struct {
	mysqlService *service.MySQLService
}

// NewHandler 创建处理器实例
func NewHandler() *Handler {
	return &Handler{
		mysqlService: service.NewMySQLService(),
	}
}

// SheetMeta 获取表结构
func (h *Handler) SheetMeta(c *gin.Context) {
	var req models.SheetMetaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "请求参数错误: " + err.Error(),
		})
		return
	}

	// 调用服务层
	data, err := h.mysqlService.GetSheetMeta(&req)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取表结构失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: data,
	})
}

// Records 获取表记录
func (h *Handler) Records(c *gin.Context) {
	var req models.RecordsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "请求参数错误: " + err.Error(),
		})
		return
	}

	// 调用服务层
	data, err := h.mysqlService.GetRecords(&req)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取表记录失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: data,
	})
}

// GetDatabases 获取数据库列表(前端辅助接口)
func (h *Handler) GetDatabases(c *gin.Context) {
	var config models.MySQLConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "请求参数错误: " + err.Error(),
		})
		return
	}

	databases, err := h.mysqlService.GetDatabases(&config)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取数据库列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: databases,
	})
}

// GetTables 获取数据表列表(前端辅助接口)
func (h *Handler) GetTables(c *gin.Context) {
	var config models.MySQLConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "请求参数错误: " + err.Error(),
		})
		return
	}

	tables, err := h.mysqlService.GetTables(&config)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取数据表列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: tables,
	})
}

// GetFields 获取表字段(前端辅助接口)
func (h *Handler) GetFields(c *gin.Context) {
	var config models.MySQLConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "请求参数错误: " + err.Error(),
		})
		return
	}

	fields, err := h.mysqlService.GetFields(&config)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取字段列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: fields,
	})
}

// PreviewSQL 预览SQL执行结果（获取字段列表）
func (h *Handler) PreviewSQL(c *gin.Context) {
	var config models.MySQLConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "请求参数错误: " + err.Error(),
		})
		return
	}

	if config.CustomSQL == "" {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "SQL语句不能为空",
		})
		return
	}

	fields, err := h.mysqlService.PreviewSQL(&config)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "SQL执行失败: " + err.Error(),
		})
		return
	}

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
