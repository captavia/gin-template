package wrap

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"template/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/samber/mo"
)

func Wrap0[Res any](h func(c *gin.Context) mo.Result[Res]) gin.HandlerFunc {
	return func(c *gin.Context) {
		handleResultWithParams(c, h(c))
	}
}

func Wrap1[T1 any, PT1 interface {
	*T1
	Extractor
}, Res any](h func(c *gin.Context, t1 PT1) mo.Result[Res]) gin.HandlerFunc {
	return func(c *gin.Context) {
		t1 := new(T1)
		if err := PT1(t1).Extract(c); err != nil {
			handleExtractError(c, t1, err)
			return
		}
		handleResultWithParams(c, h(c, t1), t1)
	}
}

func Wrap2[T1, T2 any, PT1 interface {
	*T1
	Extractor
}, PT2 interface {
	*T2
	Extractor
}, Res any](h func(c *gin.Context, t1 PT1, t2 PT2) mo.Result[Res]) gin.HandlerFunc {
	return func(c *gin.Context) {
		t1 := new(T1)
		t2 := new(T2)
		if err := PT1(t1).Extract(c); err != nil {
			handleExtractError(c, t1, err)
			return
		}
		if err := PT2(t2).Extract(c); err != nil {
			handleExtractError(c, t2, err)
			return
		}
		handleResultWithParams(c, h(c, t1, t2), t1, t2)
	}
}

func Wrap3[T1, T2, T3 any, PT1 interface {
	*T1
	Extractor
}, PT2 interface {
	*T2
	Extractor
}, PT3 interface {
	*T3
	Extractor
}, Res any](h func(c *gin.Context, t1 PT1, t2 PT2, t3 PT3) mo.Result[Res]) gin.HandlerFunc {
	return func(c *gin.Context) {
		t1 := new(T1)
		t2 := new(T2)
		t3 := new(T3)
		if err := PT1(t1).Extract(c); err != nil {
			handleExtractError(c, t1, err)
			return
		}
		if err := PT2(t2).Extract(c); err != nil {
			handleExtractError(c, t2, err)
			return
		}
		if err := PT3(t3).Extract(c); err != nil {
			handleExtractError(c, t3, err)
			return
		}
		handleResultWithParams(c, h(c, t1, t2, t3), t1, t2, t3)
	}
}

func Wrap4[T1, T2, T3, T4 any, PT1 interface {
	*T1
	Extractor
}, PT2 interface {
	*T2
	Extractor
}, PT3 interface {
	*T3
	Extractor
}, PT4 interface {
	*T4
	Extractor
}, Res any](h func(c *gin.Context, t1 PT1, t2 PT2, t3 PT3, t4 PT4) mo.Result[Res]) gin.HandlerFunc {
	return func(c *gin.Context) {
		t1, t2, t3, t4 := new(T1), new(T2), new(T3), new(T4)
		if err := PT1(t1).Extract(c); err != nil {
			handleExtractError(c, t1, err)
			return
		}
		if err := PT2(t2).Extract(c); err != nil {
			handleExtractError(c, t2, err)
			return
		}
		if err := PT3(t3).Extract(c); err != nil {
			handleExtractError(c, t3, err)
			return
		}
		if err := PT4(t4).Extract(c); err != nil {
			handleExtractError(c, t4, err)
			return
		}
		handleResultWithParams(c, h(c, t1, t2, t3, t4), t1, t2, t3, t4)
	}
}

func Wrap5[T1, T2, T3, T4, T5 any, PT1 interface {
	*T1
	Extractor
}, PT2 interface {
	*T2
	Extractor
}, PT3 interface {
	*T3
	Extractor
}, PT4 interface {
	*T4
	Extractor
}, PT5 interface {
	*T5
	Extractor
}, Res any](h func(c *gin.Context, t1 PT1, t2 PT2, t3 PT3, t4 PT4, t5 PT5) mo.Result[Res]) gin.HandlerFunc {
	return func(c *gin.Context) {
		t1, t2, t3, t4, t5 := new(T1), new(T2), new(T3), new(T4), new(T5)
		if err := PT1(t1).Extract(c); err != nil {
			handleExtractError(c, t1, err)
			return
		}
		if err := PT2(t2).Extract(c); err != nil {
			handleExtractError(c, t2, err)
			return
		}
		if err := PT3(t3).Extract(c); err != nil {
			handleExtractError(c, t3, err)
			return
		}
		if err := PT4(t4).Extract(c); err != nil {
			handleExtractError(c, t4, err)
			return
		}
		if err := PT5(t5).Extract(c); err != nil {
			handleExtractError(c, t5, err)
			return
		}
		handleResultWithParams(c, h(c, t1, t2, t3, t4, t5), t1, t2, t3, t4, t5)
	}
}

func handleResultWithParams[Res any](c *gin.Context, result mo.Result[Res], params ...any) {
	if result.IsError() {
		err := result.Error()
		if appErr, ok := errors.AsType[*utils.Response](err); ok {
			c.JSON(http.StatusOK, appErr)
		} else {
			c.JSON(http.StatusOK, utils.Err(utils.CodeError, err))
		}
		c.Abort()
		return
	}
	data := result.MustGet()
	c.JSON(http.StatusOK, utils.Ok(data))
}

// for log
func formatParams(params []any) string {
	if len(params) == 0 {
		return "[]"
	}
	var result strings.Builder
	result.WriteString("[")
	for i, p := range params {
		if i > 0 {
			result.WriteString(", ")
		}
		result.WriteString(fmt.Sprintf("%v", p))
	}
	result.WriteString("]")
	return result.String()
}

func handleExtractError(c *gin.Context, param any, err error) {
	c.JSON(http.StatusOK, utils.Err(utils.CodeInvalidParameter, fmt.Errorf("request parse error: %w", err)))
	c.Abort()
}
