package wrap

import (
	"fmt"
	"reflect"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

var maxOutputLen atomic.Int32

func init() {
	maxOutputLen.Store(500) // 默认值
}

func SetMaxOutputLen(length int) {
	maxOutputLen.Store(int32(length))
}

func smartString(data any) string {
	full := fmt.Sprintf("%+v", data)
	maxLen := int(maxOutputLen.Load())
	if len(full) <= maxLen {
		return full
	}

	// 内容过长，返回摘要
	t := reflect.TypeOf(data)
	if t.Kind() == reflect.Pointer {
		if reflect.ValueOf(data).IsNil() {
			return fmt.Sprintf("<%s: nil>", t)
		}
		t = t.Elem()
	}

	if t.Kind() == reflect.Struct {
		return fmt.Sprintf("<%s with %d fields, output too long (%d chars)>", t.Name(), t.NumField(), len(full))
	}

	return fmt.Sprintf("<%s, output too long (%d chars)>", t, len(full))
}

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
	if e := c.ShouldBindUri(&p.Data); e != nil {
		return e
	}
	return applyValidators(&p.Data)
}

func (p *Path[T]) String() string { return smartString(p.Data) }

type Query[T any] struct{ Data T }

func (q *Query[T]) Extract(c *gin.Context) error {
	applyDefaults(&q.Data)
	if e := c.ShouldBindWith(&q.Data, binding.Query); e != nil {
		return e
	}
	return applyValidators(&q.Data)
}

func (q *Query[T]) String() string { return smartString(q.Data) }

type JSON[T any] struct{ Data T }

func (j *JSON[T]) Extract(c *gin.Context) error {
	applyDefaults(&j.Data)
	if e := c.ShouldBindJSON(&j.Data); e != nil {
		return e
	}
	return applyValidators(&j.Data)
}

func (j *JSON[T]) String() string { return smartString(j.Data) }

type File[T any] struct{ Data T }

func (f *File[T]) Extract(c *gin.Context) error {
	applyDefaults(&f.Data)
	if e := c.ShouldBindWith(&f.Data, binding.FormMultipart); e != nil {
		return e
	}
	return applyValidators(&f.Data)
}

func (f *File[T]) String() string { return smartString(f.Data) }
