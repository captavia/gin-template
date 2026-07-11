package wrap

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type Extractor interface {
	Extract(c *gin.Context) error
}

// Defaulter 由请求结构体实现，用于在绑定前填充默认值。
// 绑定顺序与 json 填充结构体再 Unmarshal 一致：先调用 Default 赋默认值，
// 再由请求参数覆盖其中出现的字段，缺失的字段保留默认值。
type Defaulter interface {
	Default()
}

// applyDefaults 检测 data 是否实现了 Defaulter，若是则调用其 Default 方法。
func applyDefaults(data any) {
	if d, ok := data.(Defaulter); ok {
		d.Default()
	}
}

type Path[T any] struct{ Data T }

func (p *Path[T]) Extract(c *gin.Context) error {
	applyDefaults(&p.Data)
	return c.ShouldBindUri(&p.Data)
}

type Query[T any] struct{ Data T }

func (q *Query[T]) Extract(c *gin.Context) error {
	applyDefaults(&q.Data)
	return c.ShouldBindWith(&q.Data, binding.Query)
}

type JSON[T any] struct{ Data T }

func (j *JSON[T]) Extract(c *gin.Context) error {
	applyDefaults(&j.Data)
	return c.ShouldBindJSON(&j.Data)
}

type File[T any] struct{ Data T }

func (f *File[T]) Extract(c *gin.Context) error {
	applyDefaults(&f.Data)
	return c.ShouldBindWith(&f.Data, binding.FormMultipart)
}

type Auth[T any] struct {
	Data T
}

var GlobalAuthContextKey = "DEFAULT_USER_CLAIMS"

func (a *Auth[T]) Extract(c *gin.Context) error {
	val, exists := c.Get(GlobalAuthContextKey)
	if !exists {
		return errors.New("unauthorized: missing auth context")
	}
	data, ok := val.(T)
	if !ok {
		return errors.New("internal server error: auth claims type mismatch")
	}

	a.Data = data
	return nil
}

func SetAuthContextKey(key string) {
	GlobalAuthContextKey = key
}
