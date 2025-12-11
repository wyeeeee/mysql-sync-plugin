package handler

import (
	"mysql-sync-plugin/logger"
	"mysql-sync-plugin/models"
	"mysql-sync-plugin/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DatasourceHandler 数据源管理处理器
type DatasourceHandler struct {
	datasourceService *service.DatasourceService
	log               *logger.Logger
}

// NewDatasourceHandler 创建数据源管理处理器
func NewDatasourceHandler(datasourceService *service.DatasourceService) *DatasourceHandler {
	return &DatasourceHandler{
		datasourceService: datasourceService,
		log:               logger.New("datasource-handler"),
	}
}

// CreateDatasource 创建数据源
func (h *DatasourceHandler) CreateDatasource(c *gin.Context) {
	var req models.CreateDatasourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeAuthFailed,
			Msg:  "未找到用户信息",
		})
		return
	}

	ds, err := h.datasourceService.CreateDatasource(&req, userID.(int64))
	if err != nil {
		h.log.Errorf("创建数据源", "创建数据源失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "创建数据源失败: " + err.Error(),
		})
		return
	}

	h.log.Infof("创建数据源", "成功创建数据源: %s", ds.Name)
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: ds,
	})
}

// GetDatasource 获取数据源详情
func (h *DatasourceHandler) GetDatasource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的数据源ID",
		})
		return
	}

	ds, err := h.datasourceService.GetDatasourceByID(id)
	if err != nil {
		h.log.Errorf("获取数据源", "获取数据源失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取数据源失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: ds,
	})
}

// ListDatasources 获取数据源列表
func (h *DatasourceHandler) ListDatasources(c *gin.Context) {
	var query models.DatasourceQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	datasources, total, err := h.datasourceService.ListDatasources(&query)
	if err != nil {
		h.log.Errorf("获取数据源列表", "获取数据源列表失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取数据源列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: gin.H{
			"list":     datasources,
			"total":    total,
			"page":     query.Page,
			"pageSize": query.PageSize,
		},
	})
}

// UpdateDatasource 更新数据源
func (h *DatasourceHandler) UpdateDatasource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的数据源ID",
		})
		return
	}

	var req models.UpdateDatasourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	ds, err := h.datasourceService.UpdateDatasource(id, &req)
	if err != nil {
		h.log.Errorf("更新数据源", "更新数据源失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "更新数据源失败: " + err.Error(),
		})
		return
	}

	h.log.Infof("更新数据源", "成功更新数据源: %s", ds.Name)
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: ds,
	})
}

// DeleteDatasource 删除数据源
func (h *DatasourceHandler) DeleteDatasource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的数据源ID",
		})
		return
	}

	if err := h.datasourceService.DeleteDatasource(id); err != nil {
		h.log.Errorf("删除数据源", "删除数据源失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "删除数据源失败: " + err.Error(),
		})
		return
	}

	h.log.Infof("删除数据源", "成功删除数据源ID: %d", id)
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Msg:  "删除成功",
	})
}

// TestConnection 测试数据源连接
func (h *DatasourceHandler) TestConnection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的数据源ID",
		})
		return
	}

	if err := h.datasourceService.TestConnection(id); err != nil {
		h.log.Errorf("测试连接", "测试连接失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "连接测试失败: " + err.Error(),
		})
		return
	}

	h.log.Infof("测试连接", "数据源ID %d 连接测试成功", id)
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Msg:  "连接测试成功",
	})
}

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

// GetFieldMappings 获取字段映射
func (h *DatasourceHandler) GetFieldMappings(c *gin.Context) {
	idStr := c.Param("id")
	tableID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的表配置ID",
		})
		return
	}

	mappings, err := h.datasourceService.GetFieldMappings(tableID)
	if err != nil {
		h.log.Errorf("获取字段映射", "获取字段映射失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取字段映射失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: mappings,
	})
}

// UpdateFieldMappings 更新字段映射
func (h *DatasourceHandler) UpdateFieldMappings(c *gin.Context) {
	idStr := c.Param("id")
	tableID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的表配置ID",
		})
		return
	}

	var req struct {
		FieldMappings []models.FieldMapping `json:"fieldMappings" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.datasourceService.UpdateFieldMappings(tableID, req.FieldMappings); err != nil {
		h.log.Errorf("更新字段映射", "更新字段映射失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "更新字段映射失败: " + err.Error(),
		})
		return
	}

	h.log.Infof("更新字段映射", "成功更新表ID %d 的字段映射", tableID)
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Msg:  "更新成功",
	})
}
