package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	"template/internal/api/handler"
	"template/internal/middleware"
)

func SetupRouter(injector *do.Injector) *gin.Engine {
	r := gin.Default()

	// 跨域处理，适应前后端分离和 SSE API
	r.Use(cors.Default())

	apiV1 := r.Group("/api/v1")

	// 从 samber/do 依赖容器中获取实例化的 Handlers
	authHandler := do.MustInvoke[*handler.AuthHandler](injector)
	permHandler := do.MustInvoke[*handler.PermissionHandler](injector)

	// -- 认证系统 --
	apiV1.POST("/auth/register", authHandler.Register)
	apiV1.POST("/auth/login", authHandler.Login)

	// -- 权限管理 (仅限管理员，此处先用 AuthMiddleware 保护) --
	// 实际生产中，可以先通过数据库手动给某个用户分配超级管理员角色
	permissions := apiV1.Group("/permissions")
	permissions.Use(middleware.AuthMiddleware(), middleware.CasbinMiddleware(injector))
	{
		permissions.POST("/role", permHandler.AddRoleForUser)
		permissions.POST("/policy", permHandler.AddPermissionForRole)
		permissions.GET("/user/roles", permHandler.GetRolesForUser)
	}
	return r
}
