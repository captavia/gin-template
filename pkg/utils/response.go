package utils

import (
	"net/http"

	"github.com/samber/mo"
)

type JsonResponse struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Data    any    `json:"data"`
}

func (e *JsonResponse) Error() string {
	if e == nil {
		return EmptyMessage
	}
	return e.Message
}

func Ok(data any) *JsonResponse {
	return &JsonResponse{Code: CodeSuccess, Message: SuccessMessage, Data: data}
}

func Err(code int, msg error) *JsonResponse {
	if msg == nil {
		return &JsonResponse{Code: code, Message: EmptyMessage, Data: EmptyStruct}
	}
	return &JsonResponse{Code: code, Message: msg.Error(), Data: EmptyStruct}
}

type HTMLResponse struct {
	Code       int
	Body       string
	Template   string
	Data       any
	IsTemplate bool
}

func (h *HTMLResponse) Error() string {
	if h == nil {
		return EmptyMessage
	}
	if h.IsTemplate || h.Template != "" {
		return h.Template
	}
	return h.Body
}

func HTML[T any](body string) mo.Result[T] {
	return mo.Err[T](&HTMLResponse{
		Code: http.StatusOK,
		Body: body,
	})
}

func HTMLTemplate[T any](name string, data any) mo.Result[T] {
	return mo.Err[T](&HTMLResponse{
		Code:       http.StatusOK,
		Template:   name,
		Data:       data,
		IsTemplate: true,
	})
}

type RedirectResponse struct {
	Code     int
	Location string
}

func (r *RedirectResponse) Error() string {
	if r == nil {
		return EmptyMessage
	}
	return r.Location
}

func Redirect[T any](location string) mo.Result[T] {
	return mo.Err[T](&RedirectResponse{
		Code:     http.StatusFound,
		Location: location,
	})
}

const (
	EmptyMessage   = ""
	SuccessMessage = "ok"
)

var (
	EmptyStruct = struct{}{}
)

const (
	CodeSuccess int = 0
	CodeError       = iota
	CodeInvalidIdentifier
	CodeInvalidParameter
	CodeInvalidUsernameOrPassword
	CodeNoPermission
	CodeInternalError
)
