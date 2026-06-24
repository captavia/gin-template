package wrap

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type Extractor interface {
	Extract(c *gin.Context) error
}

type Path[T any] struct{ Data T }

func (p *Path[T]) Extract(c *gin.Context) error { return c.ShouldBindUri(&p.Data) }

type Query[T any] struct{ Data T }

func (q *Query[T]) Extract(c *gin.Context) error { return c.ShouldBindWith(&q.Data, binding.Query) }

type JSON[T any] struct{ Data T }

func (j *JSON[T]) Extract(c *gin.Context) error { return c.ShouldBindJSON(&j.Data) }

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
