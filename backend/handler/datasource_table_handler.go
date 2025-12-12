package handler

import (
	"mysql-sync-plugin/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateDatasourceTable 创建数据源表配置
func (h *DatasourceHandler) CreateDatasourceTable(c *gin.Context) {
	idStr := c.Param("id")
	datasourceID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的数据源ID",
		})
		return
	}

	var req models.CreateDatasourceTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	table, err := h.datasourceService.CreateDatasourceTable(datasourceID, &req)
	if err != nil {
		h.log.Errorf("创建表配置", "创建表配置失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "创建表配置失败: " + err.Error(),
		})
		return
	}

	h.log.Infof("创建表配置", "成功创建表配置: %s", table.TableName)
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: table,
	})
}

// ListDatasourceTables 获取数据源的表配置列表
func (h *DatasourceHandler) ListDatasourceTables(c *gin.Context) {
	idStr := c.Param("id")
	datasourceID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的数据源ID",
		})
		return
	}

	tables, err := h.datasourceService.ListDatasourceTables(datasourceID)
	if err != nil {
		h.log.Errorf("获取表配置列表", "获取表配置列表失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取表配置列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: tables,
	})
}

// GetDatasourceTable 获取表配置详情
func (h *DatasourceHandler) GetDatasourceTable(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的表配置ID",
		})
		return
	}

	table, err := h.datasourceService.GetDatasourceTableByID(id)
	if err != nil {
		h.log.Errorf("获取表配置", "获取表配置失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取表配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: table,
	})
}

// UpdateDatasourceTable 更新表配置
func (h *DatasourceHandler) UpdateDatasourceTable(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的表配置ID",
		})
		return
	}

	var req models.UpdateDatasourceTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	table, err := h.datasourceService.UpdateDatasourceTable(id, &req)
	if err != nil {
		h.log.Errorf("更新表配置", "更新表配置失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "更新表配置失败: " + err.Error(),
		})
		return
	}

	h.log.Infof("更新表配置", "成功更新表配置: %s", table.TableName)
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: table,
	})
}

// DeleteDatasourceTable 删除表配置
func (h *DatasourceHandler) DeleteDatasourceTable(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的表配置ID",
		})
		return
	}

	if err := h.datasourceService.DeleteDatasourceTable(id); err != nil {
		h.log.Errorf("删除表配置", "删除表配置失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "删除表配置失败: " + err.Error(),
		})
		return
	}

	h.log.Infof("删除表配置", "成功删除表配置ID: %d", id)
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Msg:  "删除成功",
	})
}

// BatchCreateDatasourceTables 批量创建数据源表配置
func (h *DatasourceHandler) BatchCreateDatasourceTables(c *gin.Context) {
	idStr := c.Param("id")
	datasourceID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的数据源ID",
		})
		return
	}

	var req models.BatchCreateDatasourceTablesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	tables, err := h.datasourceService.BatchCreateDatasourceTables(datasourceID, &req)
	if err != nil {
		h.log.Errorf("批量创建表配置", "批量创建表配置失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "批量创建表配置失败: " + err.Error(),
		})
		return
	}

	h.log.Infof("批量创建表配置", "成功创建 %d 个表配置", len(tables))
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: tables,
	})
}
