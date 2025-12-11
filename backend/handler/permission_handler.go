package handler

import (
	"mysql-sync-plugin/logger"
	"mysql-sync-plugin/models"
	"mysql-sync-plugin/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PermissionHandler 权限管理处理器
type PermissionHandler struct {
	permissionService *service.PermissionService
	log               *logger.Logger
}

// NewPermissionHandler 创建权限管理处理器
func NewPermissionHandler(permissionService *service.PermissionService) *PermissionHandler {
	return &PermissionHandler{
		permissionService: permissionService,
		log:               logger.New("permission-handler"),
	}
}

// GrantDatasourcePermissions 授予数据源权限
func (h *PermissionHandler) GrantDatasourcePermissions(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的用户ID",
		})
		return
	}

	var req models.GrantDatasourcePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.permissionService.GrantDatasourcePermissions(userID, req.DatasourceIDs); err != nil {
		h.log.Errorf("授予数据源权限", "授予数据源权限失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "授予数据源权限失败: " + err.Error(),
		})
		return
	}

	h.log.Infof("授予数据源权限", "成功为用户ID %d 授予 %d 个数据源权限", userID, len(req.DatasourceIDs))
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Msg:  "授予成功",
	})
}

// RevokeDatasourcePermission 撤销数据源权限
func (h *PermissionHandler) RevokeDatasourcePermission(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的用户ID",
		})
		return
	}

	dsIDStr := c.Param("dsId")
	datasourceID, err := strconv.ParseInt(dsIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的数据源ID",
		})
		return
	}

	if err := h.permissionService.RevokeDatasourcePermissions(userID, []int64{datasourceID}); err != nil {
		h.log.Errorf("撤销数据源权限", "撤销数据源权限失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "撤销数据源权限失败: " + err.Error(),
		})
		return
	}

	h.log.Infof("撤销数据源权限", "成功撤销用户ID %d 对数据源ID %d 的权限", userID, datasourceID)
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Msg:  "撤销成功",
	})
}

// RevokeDatasourcePermissions 批量撤销数据源权限
func (h *PermissionHandler) RevokeDatasourcePermissions(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的用户ID",
		})
		return
	}

	var req models.RevokeDatasourcePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.permissionService.RevokeDatasourcePermissions(userID, req.DatasourceIDs); err != nil {
		h.log.Errorf("批量撤销数据源权限", "批量撤销数据源权限失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "批量撤销数据源权限失败: " + err.Error(),
		})
		return
	}

	h.log.Infof("批量撤销数据源权限", "成功撤销用户ID %d 的 %d 个数据源权限", userID, len(req.DatasourceIDs))
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Msg:  "撤销成功",
	})
}

// ListUserDatasources 获取用户数据源列表
func (h *PermissionHandler) ListUserDatasources(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的用户ID",
		})
		return
	}

	datasources, err := h.permissionService.ListUserDatasources(userID)
	if err != nil {
		h.log.Errorf("获取用户数据源", "获取用户数据源失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取用户数据源失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: datasources,
	})
}

// ListAllDatasourcesWithPermission 获取所有数据源及权限状态
func (h *PermissionHandler) ListAllDatasourcesWithPermission(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的用户ID",
		})
		return
	}

	datasources, err := h.permissionService.ListAllDatasourcesWithPermission(userID)
	if err != nil {
		h.log.Errorf("获取数据源权限状态", "获取数据源权限状态失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取数据源权限状态失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: datasources,
	})
}

// GrantTablePermissions 授予表权限
func (h *PermissionHandler) GrantTablePermissions(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的用户ID",
		})
		return
	}

	var req models.GrantTablePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.permissionService.GrantTablePermissions(userID, req.TableIDs); err != nil {
		h.log.Errorf("授予表权限", "授予表权限失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "授予表权限失败: " + err.Error(),
		})
		return
	}

	h.log.Infof("授予表权限", "成功为用户ID %d 授予 %d 个表权限", userID, len(req.TableIDs))
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Msg:  "授予成功",
	})
}

// RevokeTablePermission 撤销表权限
func (h *PermissionHandler) RevokeTablePermission(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的用户ID",
		})
		return
	}

	tableIDStr := c.Param("tableId")
	tableID, err := strconv.ParseInt(tableIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的表ID",
		})
		return
	}

	if err := h.permissionService.RevokeTablePermissions(userID, []int64{tableID}); err != nil {
		h.log.Errorf("撤销表权限", "撤销表权限失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "撤销表权限失败: " + err.Error(),
		})
		return
	}

	h.log.Infof("撤销表权限", "成功撤销用户ID %d 对表ID %d 的权限", userID, tableID)
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Msg:  "撤销成功",
	})
}

// RevokeTablePermissions 批量撤销表权限
func (h *PermissionHandler) RevokeTablePermissions(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的用户ID",
		})
		return
	}

	var req models.RevokeTablePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.permissionService.RevokeTablePermissions(userID, req.TableIDs); err != nil {
		h.log.Errorf("批量撤销表权限", "批量撤销表权限失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "批量撤销表权限失败: " + err.Error(),
		})
		return
	}

	h.log.Infof("批量撤销表权限", "成功撤销用户ID %d 的 %d 个表权限", userID, len(req.TableIDs))
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Msg:  "撤销成功",
	})
}

// ListUserTables 获取用户表列表
func (h *PermissionHandler) ListUserTables(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的用户ID",
		})
		return
	}

	datasourceIDStr := c.Query("datasourceId")
	if datasourceIDStr == "" {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "缺少数据源ID参数",
		})
		return
	}

	datasourceID, err := strconv.ParseInt(datasourceIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的数据源ID",
		})
		return
	}

	tables, err := h.permissionService.ListUserTables(userID, datasourceID)
	if err != nil {
		h.log.Errorf("获取用户表", "获取用户表失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取用户表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: tables,
	})
}

// ListAllTablesWithPermission 获取所有表及权限状态
func (h *PermissionHandler) ListAllTablesWithPermission(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的用户ID",
		})
		return
	}

	datasourceIDStr := c.Query("datasourceId")
	if datasourceIDStr == "" {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "缺少数据源ID参数",
		})
		return
	}

	datasourceID, err := strconv.ParseInt(datasourceIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的数据源ID",
		})
		return
	}

	tables, err := h.permissionService.ListAllTablesWithPermission(userID, datasourceID)
	if err != nil {
		h.log.Errorf("获取表权限状态", "获取表权限状态失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取表权限状态失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: tables,
	})
}
