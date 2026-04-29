package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"template/pkg/utils"
)

func AuthMiddleware() gin.HandlerFunc {
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

		claims, err := utils.ParseToken(parts[1])
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
