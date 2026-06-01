package middleware

import (
	"errors"
	"net/http"
	"strings"
	"template/internal/service"

	"template/pkg/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(i do.Injector) gin.HandlerFunc {
	jwtService := do.MustInvoke[*service.JwtService](i)
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, utils.Err(utils.CodeInvalidIdentifier, errors.New("Authorization header is required ")))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, utils.Err(utils.CodeInvalidIdentifier, errors.New("Authorization header format must be Bearer {token} ")))
			c.Abort()
			return
		}

		claims, err := jwtService.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, utils.Err(utils.CodeInvalidIdentifier, err))
			c.Abort()
			return
		}

		// 将解析出的 userID 放入上下文供业务层读取
		c.Set("claims", claims)
		c.Next()
	}
}
