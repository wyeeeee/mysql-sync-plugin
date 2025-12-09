package handler

import (
	"mysql-sync-plugin/logger"
	"mysql-sync-plugin/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminHandler 管理后台处理器
type AdminHandler struct {
	log *logger.Logger
}

// NewAdminHandler 创建管理后台处理器
func NewAdminHandler() *AdminHandler {
	return &AdminHandler{
		log: logger.New("admin"),
	}
}

// GetLogs 获取日志列表
func (h *AdminHandler) GetLogs(c *gin.Context) {
	var query logger.LogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	store := logger.GetStore()
	logs, total, err := store.Query(&query)
	if err != nil {
		h.log.Errorf("查询日志", "查询日志失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "查询日志失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: gin.H{
			"list":     logs,
			"total":    total,
			"page":     query.Page,
			"pageSize": query.PageSize,
		},
	})
}

// GetLogStats 获取日志统计
func (h *AdminHandler) GetLogStats(c *gin.Context) {
	store := logger.GetStore()
	stats, err := store.GetStats()
	if err != nil {
		h.log.Errorf("统计日志", "获取统计失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取统计失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: stats,
	})
}

// CleanLogs 清理旧日志
func (h *AdminHandler) CleanLogs(c *gin.Context) {
	var req struct {
		Days int `json:"days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Days <= 0 {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "请指定有效的天数",
		})
		return
	}

	store := logger.GetStore()
	affected, err := store.CleanOldLogs(req.Days)
	if err != nil {
		h.log.Errorf("清理日志", "清理日志失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "清理日志失败: " + err.Error(),
		})
		return
	}

	h.log.Infof("清理日志", "清理了 %d 条 %d 天前的日志", affected, req.Days)
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: gin.H{
			"affected": affected,
		},
	})
}

// GetSystemInfo 获取系统信息
func (h *AdminHandler) GetSystemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: gin.H{
			"service": "mysql-sync-plugin",
			"version": "1.0.0",
		},
	})
}
