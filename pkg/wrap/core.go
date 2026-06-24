package wrap

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
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

func Wrap1[T1 Extractor, Res any](h func(t1 *T1) mo.Result[Res]) gin.HandlerFunc {
	return func(c *gin.Context) {
		var t1 T1
		if err := t1.Extract(c); err != nil {
			handleExtractError(c, err)
			return
		}
		handleResult(c, h(&t1))
	}
}

func Wrap2[T1, T2 Extractor, Res any](h func(t1 *T1, t2 *T2) mo.Result[Res]) gin.HandlerFunc {
	return func(c *gin.Context) {
		var t1 T1
		var t2 T2
		if err := t1.Extract(c); err != nil {
			handleExtractError(c, err)
			return
		}
		if err := t2.Extract(c); err != nil {
			handleExtractError(c, err)
			return
		}
		handleResult(c, h(&t1, &t2))
	}
}

func Wrap3[T1, T2, T3 Extractor, Res any](h func(t1 *T1, t2 *T2, t3 *T3) mo.Result[Res]) gin.HandlerFunc {
	return func(c *gin.Context) {
		var t1 T1
		var t2 T2
		var t3 T3
		if err := t1.Extract(c); err != nil {
			handleExtractError(c, err)
			return
		}
		if err := t2.Extract(c); err != nil {
			handleExtractError(c, err)
			return
		}
		if err := t3.Extract(c); err != nil {
			handleExtractError(c, err)
			return
		}
		handleResult(c, h(&t1, &t2, &t3))
	}
}

func Wrap4[T1, T2, T3, T4 Extractor, Res any](h func(t1 *T1, t2 *T2, t3 *T3, t4 *T4) mo.Result[Res]) gin.HandlerFunc {
	return func(c *gin.Context) {
		var t1 T1
		var t2 T2
		var t3 T3
		var t4 T4
		if err := t1.Extract(c); err != nil {
			handleExtractError(c, err)
			return
		}
		if err := t2.Extract(c); err != nil {
			handleExtractError(c, err)
			return
		}
		if err := t3.Extract(c); err != nil {
			handleExtractError(c, err)
			return
		}
		if err := t4.Extract(c); err != nil {
			handleExtractError(c, err)
			return
		}
		handleResult(c, h(&t1, &t2, &t3, &t4))
	}
}

func Wrap5[T1, T2, T3, T4, T5 Extractor, Res any](h func(t1 *T1, t2 *T2, t3 *T3, t4 *T4, t5 *T5) mo.Result[Res]) gin.HandlerFunc {
	return func(c *gin.Context) {
		var t1 T1
		var t2 T2
		var t3 T3
		var t4 T4
		var t5 T5
		if err := t1.Extract(c); err != nil {
			handleExtractError(c, err)
			return
		}
		if err := t2.Extract(c); err != nil {
			handleExtractError(c, err)
			return
		}
		if err := t3.Extract(c); err != nil {
			handleExtractError(c, err)
			return
		}
		if err := t4.Extract(c); err != nil {
			handleExtractError(c, err)
			return
		}
		if err := t5.Extract(c); err != nil {
			handleExtractError(c, err)
			return
		}
		handleResult(c, h(&t1, &t2, &t3, &t4, &t5))
	}
}

func Wrap(h any) gin.HandlerFunc {
	val := reflect.ValueOf(h)
	typ := val.Type()

	if typ.Kind() != reflect.Func {
		panic("core.Wrap: 传入的 handler 必须是一个函数!")
	}

	numIn := typ.NumIn()
	extractorInterface := reflect.TypeOf((*Extractor)(nil)).Elem()

	argElemTypes := make([]reflect.Type, numIn)
	for i := 0; i < numIn; i++ {
		argType := typ.In(i)
		if argType.Kind() != reflect.Ptr || !argType.Implements(extractorInterface) {
			panic("core.Wrap: handler all args must be pointer of Extractor")
		}
		argElemTypes[i] = argType.Elem()
	}

	if typ.NumOut() != 1 {
		panic("core.Wrap: handler must have one return value (mo.Result[T])")
	}
	outTyp := typ.Out(0)
	isErrorMethod, ok1 := outTyp.MethodByName("IsError")
	errorMethod, ok2 := outTyp.MethodByName("Error")
	mustGetMethod, ok3 := outTyp.MethodByName("MustGet")

	if !ok1 || !ok2 || !ok3 {
		panic("core.Wrap: handler return type must be mo.Result[T]")
	}

	return func(c *gin.Context) {
		args := make([]reflect.Value, numIn)

		for i := 0; i < numIn; i++ {
			argVal := reflect.New(argElemTypes[i])
			extractor := argVal.Interface().(Extractor)

			if err := extractor.Extract(c); err != nil {
				handleExtractError(c, err)
				return
			}
			args[i] = argVal
		}

		results := val.Call(args)
		resVal := results[0]

		isError := resVal.Method(isErrorMethod.Index).Call(nil)[0].Bool()

		if isError {
			errInterface := resVal.Method(errorMethod.Index).Call(nil)[0].Interface()
			err := errInterface.(error)

			var appErr *utils.Response
			if errors.As(err, &appErr) {
				c.JSON(appErr.Code, gin.H{"code": appErr.Code, "error": appErr.Message})
			} else {
				c.JSON(http.StatusInternalServerError, utils.Err(utils.CodeError, err))
			}
			c.Abort()
			return
		}

		data := resVal.Method(mustGetMethod.Index).Call(nil)[0].Interface()
		c.JSON(http.StatusOK, utils.Ok(data))
	}
}

func handleResult[Res any](c *gin.Context, result mo.Result[Res]) {
	if result.IsError() {
		err := result.Error()
		var appErr *utils.Response
		if errors.As(err, &appErr) {
			c.JSON(appErr.Code, gin.H{"code": appErr.Code, "error": appErr.Message})
		} else {
			c.JSON(http.StatusInternalServerError, utils.Err(utils.CodeError, err))
		}
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, utils.Ok(result.MustGet()))
}

func handleExtractError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, utils.Err(utils.CodeInvalidParameter, fmt.Errorf("request parse error: %v", err)))
	c.Abort()
}
