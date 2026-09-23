package wrap

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"template/pkg/utils"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/samber/mo"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// Test request and response types
type TestReq struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type TestRes struct {
	Message string `json:"message"`
}

type TestPathReq struct {
	ID string `uri:"id" binding:"required"`
}

type TestQueryReq struct {
	Keyword string `form:"keyword"`
}

// TestWrapTyped tests WrapTyped function
func TestWrapTyped(t *testing.T) {
	tests := []struct {
		name           string
		handler        TypedHandler[TestReq, TestRes]
		body           string
		wantStatusCode int
		wantContains   string
	}{
		{
			name: "successful request",
			handler: func(ctx *gin.Context, req *TestReq) mo.Result[TestRes] {
				return mo.Ok(TestRes{Message: fmt.Sprintf("Hello %s", req.Name)})
			},
			body:           `{"name":"John","age":30}`,
			wantStatusCode: http.StatusOK,
			wantContains:   "Hello John",
		},
		{
			name: "handler returns error",
			handler: func(ctx *gin.Context, req *TestReq) mo.Result[TestRes] {
				return mo.Err[TestRes](errors.New("internal error"))
			},
			body:           `{"name":"John","age":30}`,
			wantStatusCode: http.StatusInternalServerError,
			wantContains:   "internal error",
		},
		{
			name: "invalid JSON",
			handler: func(ctx *gin.Context, req *TestReq) mo.Result[TestRes] {
				return mo.Ok(TestRes{Message: "should not reach here"})
			},
			body:           `{invalid json}`,
			wantStatusCode: http.StatusBadRequest,
			wantContains:   "请求参数错误",
		},
		{
			name: "empty body",
			handler: func(ctx *gin.Context, req *TestReq) mo.Result[TestRes] {
				return mo.Ok(TestRes{Message: "Empty request"})
			},
			body:           "",
			wantStatusCode: http.StatusOK,
			wantContains:   "Empty request",
		},
		{
			name: "handler returns app error",
			handler: func(ctx *gin.Context, req *TestReq) mo.Result[TestRes] {
				return mo.Err[TestRes](&utils.Response{Code: http.StatusUnauthorized, Message: "unauthorized"})
			},
			body:           `{"name":"John"}`,
			wantStatusCode: http.StatusUnauthorized,
			wantContains:   "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/test", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler := WrapTyped(tt.handler)
			handler(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("got status code %v, want %v", w.Code, tt.wantStatusCode)
			}

			if !strings.Contains(w.Body.String(), tt.wantContains) {
				t.Errorf("response body %v does not contain %v", w.Body.String(), tt.wantContains)
			}
		})
	}
}

// TestWrap1 tests Wrap1 function
func TestWrap1(t *testing.T) {
	tests := []struct {
		name           string
		handler        func(c *gin.Context, req *JSON[TestReq]) mo.Result[TestRes]
		body           string
		wantStatusCode int
		wantContains   string
	}{
		{
			name: "successful request",
			handler: func(c *gin.Context, req *JSON[TestReq]) mo.Result[TestRes] {
				return mo.Ok(TestRes{Message: fmt.Sprintf("Name: %s", req.Data.Name)})
			},
			body:           `{"name":"Alice","age":25}`,
			wantStatusCode: http.StatusOK,
			wantContains:   "Alice",
		},
		{
			name: "extraction error",
			handler: func(c *gin.Context, req *JSON[TestReq]) mo.Result[TestRes] {
				return mo.Ok(TestRes{Message: "should not reach"})
			},
			body:           `{invalid}`,
			wantStatusCode: http.StatusBadRequest,
			wantContains:   "request parse error",
		},
		{
			name: "handler returns error",
			handler: func(c *gin.Context, req *JSON[TestReq]) mo.Result[TestRes] {
				return mo.Err[TestRes](errors.New("business error"))
			},
			body:           `{"name":"Bob"}`,
			wantStatusCode: http.StatusInternalServerError,
			wantContains:   "business error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/test", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler := Wrap1(tt.handler)
			handler(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("got status code %v, want %v", w.Code, tt.wantStatusCode)
			}

			if !strings.Contains(w.Body.String(), tt.wantContains) {
				t.Errorf("response body %v does not contain %v", w.Body.String(), tt.wantContains)
			}
		})
	}
}

// TestWrap2 tests Wrap2 function with two parameters
func TestWrap2(t *testing.T) {
	handler := func(c *gin.Context, path *Path[TestPathReq], query *Query[TestQueryReq]) mo.Result[TestRes] {
		return mo.Ok(TestRes{
			Message: fmt.Sprintf("ID: %s, Keyword: %s", path.Data.ID, query.Data.Keyword),
		})
	}

	tests := []struct {
		name           string
		pathID         string
		query          string
		wantStatusCode int
		wantContains   string
	}{
		{
			name:           "successful request",
			pathID:         "123",
			query:          "keyword=search",
			wantStatusCode: http.StatusOK,
			wantContains:   "ID: 123, Keyword: search",
		},
		{
			name:           "missing path param",
			pathID:         "",
			query:          "keyword=search",
			wantStatusCode: http.StatusBadRequest,
			wantContains:   "request parse error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test?"+tt.query, nil)

			if tt.pathID != "" {
				c.Params = []gin.Param{{Key: "id", Value: tt.pathID}}
			}

			wrappedHandler := Wrap2(handler)
			wrappedHandler(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("got status code %v, want %v", w.Code, tt.wantStatusCode)
			}

			if !strings.Contains(w.Body.String(), tt.wantContains) {
				t.Errorf("response body %v does not contain %v", w.Body.String(), tt.wantContains)
			}
		})
	}
}

// TestWrap3 tests Wrap3 function with three parameters
func TestWrap3(t *testing.T) {
	handler := func(c *gin.Context, p1 *JSON[TestReq], p2 *Query[TestQueryReq], p3 *Path[TestPathReq]) mo.Result[TestRes] {
		return mo.Ok(TestRes{Message: "three params"})
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/test?keyword=search", strings.NewReader(`{"name":"test","age":25}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = []gin.Param{{Key: "id", Value: "123"}}

	wrappedHandler := Wrap3(handler)
	wrappedHandler(c)

	if w.Code != http.StatusOK {
		t.Errorf("got status code %v, want %v. Response: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

// TestWrap4 tests Wrap4 function with four parameters
func TestWrap4(t *testing.T) {
	// Since we can't bind the same request body 4 times, we'll test with Query parameters
	handler := func(c *gin.Context, p1, p2, p3, p4 *Query[TestQueryReq]) mo.Result[TestRes] {
		return mo.Ok(TestRes{Message: "four params"})
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test?keyword=search", nil)

	wrappedHandler := Wrap4(handler)
	wrappedHandler(c)

	if w.Code != http.StatusOK {
		t.Errorf("got status code %v, want %v. Response: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

// TestWrap5 tests Wrap5 function with five parameters
func TestWrap5(t *testing.T) {
	// Test with Query parameters which can be reused
	handler := func(c *gin.Context, p1, p2, p3, p4, p5 *Query[TestQueryReq]) mo.Result[TestRes] {
		return mo.Ok(TestRes{Message: "five params"})
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test?keyword=search", nil)

	wrappedHandler := Wrap5(handler)
	wrappedHandler(c)

	if w.Code != http.StatusOK {
		t.Errorf("got status code %v, want %v. Response: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

// TestWrap tests the generic Wrap function with reflection
func TestWrap(t *testing.T) {
	tests := []struct {
		name           string
		handler        any
		setup          func(*gin.Context)
		body           string
		wantStatusCode int
		wantContains   string
		shouldPanic    bool
		panicContains  string
	}{
		{
			name: "valid handler with one param",
			handler: func(req *JSON[TestReq]) mo.Result[TestRes] {
				return mo.Ok(TestRes{Message: fmt.Sprintf("Hello %s", req.Data.Name)})
			},
			body:           `{"name":"World"}`,
			wantStatusCode: http.StatusOK,
			wantContains:   "Hello World",
		},
		{
			name: "valid handler with two params",
			handler: func(p1 *Path[TestPathReq], p2 *Query[TestQueryReq]) mo.Result[TestRes] {
				return mo.Ok(TestRes{Message: fmt.Sprintf("ID: %s", p1.Data.ID)})
			},
			setup: func(c *gin.Context) {
				c.Params = []gin.Param{{Key: "id", Value: "999"}}
			},
			wantStatusCode: http.StatusOK,
			wantContains:   "ID: 999",
		},
		{
			name: "handler returns error",
			handler: func(req *JSON[TestReq]) mo.Result[TestRes] {
				return mo.Err[TestRes](errors.New("test error"))
			},
			body:           `{"name":"test"}`,
			wantStatusCode: http.StatusInternalServerError,
			wantContains:   "test error",
		},
		{
			name: "handler returns app error",
			handler: func(req *JSON[TestReq]) mo.Result[TestRes] {
				return mo.Err[TestRes](&utils.Response{Code: http.StatusForbidden, Message: "forbidden"})
			},
			body:           `{"name":"test"}`,
			wantStatusCode: http.StatusForbidden,
			wantContains:   "forbidden",
		},
		{
			name:          "non-function handler panics",
			handler:       "not a function",
			shouldPanic:   true,
			panicContains: "必须是一个函数",
		},
		{
			name: "non-pointer param panics",
			handler: func(req JSON[TestReq]) mo.Result[TestRes] {
				return mo.Ok(TestRes{})
			},
			shouldPanic:   true,
			panicContains: "must be pointer of Extractor",
		},
		{
			name: "multiple return values panics",
			handler: func(req *JSON[TestReq]) (mo.Result[TestRes], error) {
				return mo.Ok(TestRes{}), nil
			},
			shouldPanic:   true,
			panicContains: "must have one return value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					r := recover()
					if r == nil {
						t.Error("expected panic but got none")
						return
					}
					panicMsg := fmt.Sprint(r)
					if !strings.Contains(panicMsg, tt.panicContains) {
						t.Errorf("panic message %v does not contain %v", panicMsg, tt.panicContains)
					}
				}()
				Wrap(tt.handler)
				return
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			if tt.body != "" {
				c.Request = httptest.NewRequest("POST", "/test", strings.NewReader(tt.body))
				c.Request.Header.Set("Content-Type", "application/json")
			} else {
				c.Request = httptest.NewRequest("GET", "/test", nil)
			}

			if tt.setup != nil {
				tt.setup(c)
			}

			handler := Wrap(tt.handler)
			handler(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("got status code %v, want %v", w.Code, tt.wantStatusCode)
			}

			if tt.wantContains != "" && !strings.Contains(w.Body.String(), tt.wantContains) {
				t.Errorf("response body %v does not contain %v", w.Body.String(), tt.wantContains)
			}
		})
	}
}

// TestHandleResult tests handleResult function
func TestHandleResult(t *testing.T) {
	tests := []struct {
		name           string
		result         mo.Result[TestRes]
		wantStatusCode int
		wantContains   string
	}{
		{
			name:           "success result",
			result:         mo.Ok(TestRes{Message: "success"}),
			wantStatusCode: http.StatusOK,
			wantContains:   "success",
		},
		{
			name:           "error result",
			result:         mo.Err[TestRes](errors.New("failed")),
			wantStatusCode: http.StatusInternalServerError,
			wantContains:   "failed",
		},
		{
			name:           "app error result",
			result:         mo.Err[TestRes](&utils.Response{Code: http.StatusNotFound, Message: "not found"}),
			wantStatusCode: http.StatusOK,
			wantContains:   "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			handleResult(c, tt.result)

			if w.Code != tt.wantStatusCode {
				t.Errorf("got status code %v, want %v", w.Code, tt.wantStatusCode)
			}

			if !strings.Contains(w.Body.String(), tt.wantContains) {
				t.Errorf("response body %v does not contain %v", w.Body.String(), tt.wantContains)
			}
		})
	}
}

// TestHandleExtractError tests handleExtractError function
func TestHandleExtractError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handleExtractError(c, errors.New("parse error"))

	if w.Code != http.StatusBadRequest {
		t.Errorf("got status code %v, want %v", w.Code, http.StatusBadRequest)
	}

	if !strings.Contains(w.Body.String(), "request parse error") {
		t.Errorf("response body %v does not contain 'request parse error'", w.Body.String())
	}

	if !strings.Contains(w.Body.String(), "parse error") {
		t.Errorf("response body %v does not contain 'parse error'", w.Body.String())
	}
}
