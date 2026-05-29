package router

import (
	"template/pkg/warp"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	"template/internal/api/handler"
)

func SetupRouter(injector *do.Injector) *gin.Engine {
	r := gin.Default()

	// 跨域处理，适应前后端分离和 SSE API
	r.Use(cors.Default())

	apiV1 := r.Group("/api/v1")

	// 从 samber/do 依赖容器中获取实例化的 Handlers
	authHandler := do.MustInvoke[*handler.AuthHandler](injector)

	// -- 认证系统 --
	apiV1.POST("/auth/register", warp.WrapTyped(authHandler.Register))
	apiV1.POST("/auth/login", warp.WrapTyped(authHandler.Login))

	return r
}
