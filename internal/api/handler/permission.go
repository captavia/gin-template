package handler

import (
	"net/http"
	"template/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	"template/internal/service"
)

type PermissionHandler struct {
	permService *service.PermissionService
}

func NewPermissionHandler(i *do.Injector) (*PermissionHandler, error) {
	return &PermissionHandler{
		permService: do.MustInvoke[*service.PermissionService](i),
	}, nil
}

type addRoleRequest struct {
	User string `json:"user" binding:"required"`
	Role string `json:"role" binding:"required"`
}

func (h *PermissionHandler) AddRoleForUser(c *gin.Context) {
	var req addRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Err(utils.CodeInvalidParameter, err))
		return
	}

	if _, err := h.permService.AddRoleForUser(req.User, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, utils.Err(utils.CodeInternalError, err))
		return
	}

	c.JSON(http.StatusOK, utils.Ok("role added successfully"))
}

type addPolicyRequest struct {
	Role   string `json:"role" binding:"required"`
	Path   string `json:"path" binding:"required"`
	Method string `json:"method" binding:"required"`
}

func (h *PermissionHandler) AddPermissionForRole(c *gin.Context) {
	var req addPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Err(utils.CodeInvalidParameter, err))
		return
	}

	if _, err := h.permService.AddPermissionForRole(req.Role, req.Path, req.Method); err != nil {
		c.JSON(http.StatusInternalServerError, utils.Err(utils.CodeInternalError, err))
		return
	}

	c.JSON(http.StatusOK, utils.Ok("permission added successfully"))
}

func (h *PermissionHandler) GetRolesForUser(c *gin.Context) {
	user := c.Query("user")
	if user == "" {
		c.JSON(http.StatusBadRequest, utils.Err(utils.CodeInvalidParameter, nil))
		return
	}

	roles, err := h.permService.GetRolesForUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Err(utils.CodeInternalError, err))
		return
	}

	c.JSON(http.StatusOK, utils.Ok(roles))
}
