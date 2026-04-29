package utils

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Data    any    `json:"data"`
}

func Ok(data any) Response {
	return Response{Code: 0, Message: SuccessMessage, Data: data}
}

func Err(code int, msg error) Response {
	return Response{Code: code, Message: msg.Error(), Data: EmptyStruct}
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
)
