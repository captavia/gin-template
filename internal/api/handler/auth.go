package handler

import (
	"net/http"
	"template/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
	"github.com/samber/mo"

	"template/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(i do.Injector) (*AuthHandler, error) {
	return &AuthHandler{
		authService: do.MustInvoke[*service.AuthService](i),
	}, nil
}

type authRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

func (h *AuthHandler) Register(c *gin.Context, req *authRequest) mo.Result[string] {
	if err := h.authService.Register(c.Request.Context(), req.Phone, req.Password); err != nil {
		return mo.Err[string](utils.Err(http.StatusConflict, err))
	}

	return mo.Ok("register successfully")
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Login(c *gin.Context, req *authRequest) mo.Result[LoginResponse] {
	token, err := h.authService.Login(c.Request.Context(), req.Phone, req.Password)
	if err != nil {
		return mo.Err[LoginResponse](utils.Err(http.StatusUnauthorized, err))
	}

	return mo.Ok(LoginResponse{Token: token})
}
