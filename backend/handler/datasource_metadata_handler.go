package handler

import (
	"mysql-sync-plugin/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetDatabaseList 获取数据源的数据库列表
func (h *DatasourceHandler) GetDatabaseList(c *gin.Context) {
	idStr := c.Param("id")
	datasourceID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的数据源ID",
		})
		return
	}

	databases, err := h.datasourceService.GetDatabaseList(datasourceID)
	if err != nil {
		h.log.Errorf("获取数据库列表", "获取数据库列表失败: %v", err)
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

// GetTableList 获取数据源指定数据库的表列表
func (h *DatasourceHandler) GetTableList(c *gin.Context) {
	idStr := c.Param("id")
	datasourceID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的数据源ID",
		})
		return
	}

	databaseName := c.Query("database")
	if databaseName == "" {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "缺少数据库名参数",
		})
		return
	}

	tables, err := h.datasourceService.GetTableList(datasourceID, databaseName)
	if err != nil {
		h.log.Errorf("获取表列表", "获取表列表失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取表列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: tables,
	})
}

// GetFieldList 获取数据源指定表的字段列表
func (h *DatasourceHandler) GetFieldList(c *gin.Context) {
	idStr := c.Param("id")
	datasourceID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的数据源ID",
		})
		return
	}

	databaseName := c.Query("database")
	tableName := c.Query("table")
	if databaseName == "" || tableName == "" {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "缺少数据库名或表名参数",
		})
		return
	}

	fields, err := h.datasourceService.GetFieldList(datasourceID, databaseName, tableName)
	if err != nil {
		h.log.Errorf("获取字段列表", "获取字段列表失败: %v", err)
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

// GetFieldListFromSQL 从自定义SQL获取字段列表
func (h *DatasourceHandler) GetFieldListFromSQL(c *gin.Context) {
	idStr := c.Param("id")
	datasourceID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的数据源ID",
		})
		return
	}

	var req struct {
		Database  string `json:"database" binding:"required"`
		CustomSQL string `json:"customSql" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	fields, err := h.datasourceService.GetFieldListFromSQL(datasourceID, req.Database, req.CustomSQL)
	if err != nil {
		h.log.Errorf("获取SQL字段列表", "获取SQL字段列表失败: %v", err)
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
