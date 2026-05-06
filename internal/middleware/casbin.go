package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/casbin/casbin/v3"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	"template/pkg/utils"
)

func CasbinMiddleware(injector *do.Injector) gin.HandlerFunc {
	enforcer := do.MustInvoke[*casbin.Enforcer](injector)

	return func(c *gin.Context) {
		claimsVal, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, utils.Err(utils.CodeInvalidIdentifier, errors.New("user not authenticated")))
			c.Abort()
			return
		}

		claims, ok := claimsVal.(*utils.Claims)
		if !ok {
			c.JSON(http.StatusInternalServerError, utils.Err(utils.CodeInternalError, errors.New("failed to parse claims")))
			c.Abort()
			return
		}

		// 2. 获取请求路径和方法
		obj := c.Request.URL.Path
		act := c.Request.Method
		sub := fmt.Sprintf("%d", claims.UserID)

		// 3. 执行权限校验
		allowed, err := enforcer.Enforce(sub, obj, act)
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.Err(utils.CodeInternalError, err))
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusForbidden, utils.Err(utils.CodeNoPermission, errors.New("permission denied")))
			c.Abort()
			return
		}

		c.Next()
	}
}
