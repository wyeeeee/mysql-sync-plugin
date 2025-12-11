package handler

import (
	"mysql-sync-plugin/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

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
