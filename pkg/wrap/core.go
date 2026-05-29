package wrap

import (
	"errors"
	"fmt"
	"net/http"
	"template/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/samber/mo"
)

type TypedHandler[Req any, Res any] func(ctx *gin.Context, req *Req) mo.Result[Res]

func WrapTyped[Req any, Res any](h TypedHandler[Req, Res]) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req

		if err := c.ShouldBind(&req); err != nil && err.Error() != "EOF" {
			c.JSON(http.StatusBadRequest, utils.Err(utils.CodeInvalidParameter, fmt.Errorf("请求参数错误: %e", err)))
			return
		}

		result := h(c, &req)

		if result.IsError() {
			err := result.Error()
			var appErr *utils.Response
			if errors.As(err, &appErr) {
				c.JSON(appErr.Code, gin.H{"code": appErr.Code, "error": appErr.Message})
			} else {
				c.JSON(http.StatusInternalServerError, utils.Err(utils.CodeError, err))
			}
			return
		}

		c.JSON(http.StatusOK, utils.Ok(result.MustGet()))
	}
}
