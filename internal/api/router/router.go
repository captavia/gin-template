package router

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	"template/internal/api/handler"
)

func SetupRouter(injector *do.Injector) *gin.Engine {
	r := gin.Default()

	// 跨域处理，适应前后端分离和 SSE API
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	apiV1 := r.Group("/api/v1")

	// 从 samber/do 依赖容器中获取实例化的 Handlers
	authHandler := do.MustInvoke[*handler.AuthHandler](injector)

	// -- 认证系统 --
	apiV1.POST("/auth/register", authHandler.Register)
	apiV1.POST("/auth/login", authHandler.Login)

	return r
}
