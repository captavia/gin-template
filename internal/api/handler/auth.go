package handler

import (
	"net/http"
	"template/pkg/utils"
	"template/pkg/wrap"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
	"github.com/samber/mo"

	"template/internal/api/service"
)

type AuthHandler struct {
	userService *service.UserService
}

func NewAuthHandler(i do.Injector) (*AuthHandler, error) {
	return &AuthHandler{
		userService: do.MustInvoke[*service.UserService](i),
	}, nil
}

type authRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

func (h *AuthHandler) Register(c *gin.Context, req *authRequest) mo.Result[string] {
	if err := h.userService.Register(c.Request.Context(), req.Phone, req.Password); err != nil {
		return mo.Err[string](utils.Err(http.StatusConflict, err))
	}

	return mo.Ok("register successfully")
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Login(req wrap.JSON[authRequest]) mo.Result[LoginResponse] {
	return mo.Ok(LoginResponse{Token: req.Data.Phone})
}
