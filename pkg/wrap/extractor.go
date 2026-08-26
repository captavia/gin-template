package wrap

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type Extractor interface {
	Extract(c *gin.Context) error
}

type Defaulter interface {
	Default()
}

type Validator interface {
	Validate() error
}

func applyDefaults(data any) {
	if d, ok := data.(Defaulter); ok {
		d.Default()
	}
}

func applyValidators(data any) error {
	if d, ok := data.(Validator); ok {
		return d.Validate()
	}
	return nil
}

type Path[T any] struct{ Data T }

func (p *Path[T]) Extract(c *gin.Context) error {
	applyDefaults(&p.Data)
	if e := applyValidators(&p.Data); e != nil {
		return e
	}
	return c.ShouldBindUri(&p.Data)
}

type Query[T any] struct{ Data T }

func (q *Query[T]) Extract(c *gin.Context) error {
	applyDefaults(&q.Data)
	if e := applyValidators(&q.Data); e != nil {
		return e
	}
	return c.ShouldBindWith(&q.Data, binding.Query)
}

type JSON[T any] struct{ Data T }

func (j *JSON[T]) Extract(c *gin.Context) error {
	applyDefaults(&j.Data)
	if e := applyValidators(&j.Data); e != nil {
		return e
	}
	return c.ShouldBindJSON(&j.Data)
}

type File[T any] struct{ Data T }

func (f *File[T]) Extract(c *gin.Context) error {
	applyDefaults(&f.Data)
	if e := applyValidators(&f.Data); e != nil {
		return e
	}
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
