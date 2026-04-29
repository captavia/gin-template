package handler

import (
	"net/http"
	"template/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	"template/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(i *do.Injector) (*AuthHandler, error) {
	return &AuthHandler{
		authService: do.MustInvoke[service.AuthService](i),
	}, nil
}

type authRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Err(utils.CodeInvalidParameter, err))
		return
	}

	if err := h.authService.Register(c.Request.Context(), req.Phone, req.Password); err != nil {
		c.JSON(http.StatusConflict, utils.Err(utils.CodeError, err))
		return
	}

	c.JSON(http.StatusOK, utils.Ok("register successfully"))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Err(utils.CodeInvalidParameter, err))
		return
	}

	token, err := h.authService.Login(c.Request.Context(), req.Phone, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.Err(utils.CodeInvalidUsernameOrPassword, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
