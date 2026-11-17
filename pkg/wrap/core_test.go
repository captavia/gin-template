package wrap

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
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

// TestWrap0 tests Wrap0 function across JSON, HTML, HTMLTemplate, and Redirect responses
func TestWrap0(t *testing.T) {
	tests := []struct {
		name            string
		handler         func(c *gin.Context) mo.Result[TestRes]
		setupEngine     func(*gin.Engine)
		wantStatusCode  int
		wantContentType string
		wantLocation    string
		wantContains    string
	}{
		{
			name: "json ok",
			handler: func(c *gin.Context) mo.Result[TestRes] {
				return mo.Ok(TestRes{Message: "pong"})
			},
			wantStatusCode:  http.StatusOK,
			wantContentType: "application/json",
			wantContains:    "pong",
		},
		{
			name: "html raw",
			handler: func(c *gin.Context) mo.Result[TestRes] {
				return utils.HTML[TestRes]("<h1>Hello</h1>")
			},
			wantStatusCode:  http.StatusOK,
			wantContentType: "text/html",
			wantContains:    "<h1>Hello</h1>",
		},
		{
			name: "html template",
			handler: func(c *gin.Context) mo.Result[TestRes] {
				return utils.HTMLTemplate[TestRes]("hello.tmpl", gin.H{"Name": "World"})
			},
			setupEngine: func(e *gin.Engine) {
				e.SetHTMLTemplate(template.Must(template.New("hello.tmpl").Parse("<h1>Hello {{.Name}}</h1>")))
			},
			wantStatusCode:  http.StatusOK,
			wantContentType: "text/html",
			wantContains:    "<h1>Hello World</h1>",
		},
		{
			name: "redirect",
			handler: func(c *gin.Context) mo.Result[TestRes] {
				return utils.Redirect[TestRes]("https://example.com/callback")
			},
			wantStatusCode: http.StatusFound,
			wantLocation:   "https://example.com/callback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, engine := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)
			if tt.setupEngine != nil {
				tt.setupEngine(engine)
			}

			Wrap0(tt.handler)(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("got status code %v, want %v", w.Code, tt.wantStatusCode)
			}
			if tt.wantContentType != "" && !strings.Contains(w.Header().Get("Content-Type"), tt.wantContentType) {
				t.Errorf("Content-Type %q does not contain %q", w.Header().Get("Content-Type"), tt.wantContentType)
			}
			if tt.wantLocation != "" && w.Header().Get("Location") != tt.wantLocation {
				t.Errorf("Location %q, want %q", w.Header().Get("Location"), tt.wantLocation)
			}
			if tt.wantContains != "" && !strings.Contains(w.Body.String(), tt.wantContains) {
				t.Errorf("response body %q does not contain %q", w.Body.String(), tt.wantContains)
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
		wantCode       int
		wantContains   string
		wantLocation   string
	}{
		{
			name: "successful request",
			handler: func(c *gin.Context, req *JSON[TestReq]) mo.Result[TestRes] {
				return mo.Ok(TestRes{Message: fmt.Sprintf("Name: %s", req.Data.Name)})
			},
			body:           `{"name":"Alice","age":25}`,
			wantStatusCode: http.StatusOK,
			wantCode:       utils.CodeSuccess,
			wantContains:   "Alice",
		},
		{
			name: "extraction error",
			handler: func(c *gin.Context, req *JSON[TestReq]) mo.Result[TestRes] {
				return mo.Ok(TestRes{Message: "should not reach"})
			},
			body:           `{invalid}`,
			wantStatusCode: http.StatusOK,
			wantCode:       utils.CodeInvalidParameter,
			wantContains:   "request parse error",
		},
		{
			name: "handler returns error",
			handler: func(c *gin.Context, req *JSON[TestReq]) mo.Result[TestRes] {
				return mo.Err[TestRes](errors.New("business error"))
			},
			body:           `{"name":"Bob"}`,
			wantStatusCode: http.StatusOK,
			wantCode:       utils.CodeError,
			wantContains:   "business error",
		},
		{
			name: "handler returns html",
			handler: func(c *gin.Context, req *JSON[TestReq]) mo.Result[TestRes] {
				return utils.HTML[TestRes](fmt.Sprintf("<p>%s</p>", req.Data.Name))
			},
			body:           `{"name":"Eve"}`,
			wantStatusCode: http.StatusOK,
			wantContains:   "<p>Eve</p>",
		},
		{
			name: "handler returns redirect",
			handler: func(c *gin.Context, req *JSON[TestReq]) mo.Result[TestRes] {
				return utils.Redirect[TestRes]("/users/" + req.Data.Name)
			},
			body:           `{"name":"charlie"}`,
			wantStatusCode: http.StatusFound,
			wantLocation:   "/users/charlie",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, engine := gin.CreateTestContext(w)
			engine.POST("/test", Wrap1(tt.handler))

			req := httptest.NewRequest("POST", "/test", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			engine.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("got status code %v, want %v", w.Code, tt.wantStatusCode)
			}

			if tt.wantLocation != "" && w.Header().Get("Location") != tt.wantLocation {
				t.Errorf("got Location %q, want %q", w.Header().Get("Location"), tt.wantLocation)
			}

			if tt.wantContains != "" && !strings.Contains(w.Body.String(), tt.wantContains) {
				t.Errorf("response body %v does not contain %v", w.Body.String(), tt.wantContains)
			}

			if tt.wantLocation == "" && strings.Contains(w.Header().Get("Content-Type"), "application/json") {
				var resp utils.JsonResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal json response: %v", err)
				}
				if resp.Code != tt.wantCode {
					t.Errorf("got json code %v, want %v", resp.Code, tt.wantCode)
				}
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
		wantCode       int
		wantContains   string
	}{
		{
			name:           "successful request",
			pathID:         "123",
			query:          "keyword=search",
			wantStatusCode: http.StatusOK,
			wantCode:       utils.CodeSuccess,
			wantContains:   "ID: 123, Keyword: search",
		},
		{
			name:           "missing path param",
			pathID:         "",
			query:          "keyword=search",
			wantStatusCode: http.StatusOK,
			wantCode:       utils.CodeInvalidParameter,
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

			var resp utils.JsonResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to unmarshal json response: %v", err)
			}
			if resp.Code != tt.wantCode {
				t.Errorf("got json code %v, want %v", resp.Code, tt.wantCode)
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

// TestHandleResultWithParams tests handleResultWithParams for JSON, HTML, HTMLTemplate, and Redirect responses
func TestHandleResultWithParams(t *testing.T) {
	t.Run("success result wraps in JsonResponse", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		handleResultWithParams(c, mo.Ok(TestRes{Message: "success"}))

		if w.Code != http.StatusOK {
			t.Fatalf("got status code %v, want %v", w.Code, http.StatusOK)
		}
		var resp utils.JsonResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if resp.Code != utils.CodeSuccess || resp.Message != utils.SuccessMessage {
			t.Fatalf("got %+v, want Code=%d Message=%q", resp, utils.CodeSuccess, utils.SuccessMessage)
		}
		if !strings.Contains(w.Body.String(), "success") {
			t.Fatalf("body %q does not contain 'success'", w.Body.String())
		}
	})

	t.Run("generic error result", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		handleResultWithParams(c, mo.Err[TestRes](errors.New("failed")))

		if w.Code != http.StatusOK {
			t.Fatalf("got status code %v, want %v", w.Code, http.StatusOK)
		}
		if !c.IsAborted() {
			t.Fatal("expected context to be aborted")
		}
		var resp utils.JsonResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if resp.Code != utils.CodeError || resp.Message != "failed" {
			t.Fatalf("got %+v, want Code=%d Message='failed'", resp, utils.CodeError)
		}
	})

	t.Run("app JsonResponse error result", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		handleResultWithParams(c, mo.Err[TestRes](&utils.JsonResponse{Code: http.StatusNotFound, Message: "not found"}))

		if w.Code != http.StatusOK {
			t.Fatalf("got status code %v, want %v", w.Code, http.StatusOK)
		}
		if !c.IsAborted() {
			t.Fatal("expected context to be aborted")
		}
		var resp utils.JsonResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if resp.Code != http.StatusNotFound || resp.Message != "not found" {
			t.Fatalf("got %+v, want Code=404 Message='not found'", resp)
		}
	})

	t.Run("HTML shorthand and custom HTMLResponse error", func(t *testing.T) {
		cases := []struct {
			name     string
			invoke   func(*gin.Context)
			wantCode int
			want     string
		}{
			{
				name: "utils.HTML shorthand",
				invoke: func(c *gin.Context) {
					handleResultWithParams(c, utils.HTML[TestRes]("<h1>Hello</h1>"))
				},
				wantCode: http.StatusOK,
				want:     "<h1>Hello</h1>",
			},
			{
				name: "custom status HTMLResponse",
				invoke: func(c *gin.Context) {
					handleResultWithParams(c, mo.Err[TestRes](&utils.HTMLResponse{Code: http.StatusCreated, Body: "<p>created</p>"}))
				},
				wantCode: http.StatusCreated,
				want:     "<p>created</p>",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest("GET", "/test", nil)

				tc.invoke(c)

				if w.Code != tc.wantCode {
					t.Fatalf("got status %v, want %v", w.Code, tc.wantCode)
				}
				if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
					t.Fatalf("Content-Type = %q, want text/html", w.Header().Get("Content-Type"))
				}
				if w.Body.String() != tc.want {
					t.Fatalf("body = %q, want %q", w.Body.String(), tc.want)
				}
			})
		}
	})

	t.Run("HTMLTemplate shorthand and custom HTMLResponse template error", func(t *testing.T) {
		tmpl := template.Must(template.New("hello.tmpl").Parse("<h1>Hello {{.Name}}</h1>"))

		cases := []struct {
			name   string
			invoke func(*gin.Context)
		}{
			{
				name: "utils.HTMLTemplate shorthand",
				invoke: func(c *gin.Context) {
					handleResultWithParams(c, utils.HTMLTemplate[TestRes]("hello.tmpl", gin.H{"Name": "World"}))
				},
			},
			{
				name: "custom HTMLResponse template",
				invoke: func(c *gin.Context) {
					handleResultWithParams(c, mo.Err[TestRes](&utils.HTMLResponse{
						Code:       http.StatusOK,
						Template:   "hello.tmpl",
						Data:       gin.H{"Name": "World"},
						IsTemplate: true,
					}))
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				c, engine := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest("GET", "/test", nil)
				engine.SetHTMLTemplate(tmpl)

				tc.invoke(c)

				if w.Code != http.StatusOK {
					t.Fatalf("got status %v, want %v", w.Code, http.StatusOK)
				}
				if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
					t.Fatalf("Content-Type = %q, want text/html", w.Header().Get("Content-Type"))
				}
				if w.Body.String() != "<h1>Hello World</h1>" {
					t.Fatalf("body = %q, want %q", w.Body.String(), "<h1>Hello World</h1>")
				}
			})
		}
	})

	t.Run("Redirect shorthand and custom code RedirectResponse error", func(t *testing.T) {
		cases := []struct {
			name         string
			invoke       func(*gin.Context)
			wantCode     int
			wantLocation string
		}{
			{
				name: "utils.Redirect shorthand",
				invoke: func(c *gin.Context) {
					handleResultWithParams(c, utils.Redirect[TestRes]("https://example.com/callback"))
				},
				wantCode:     http.StatusFound,
				wantLocation: "https://example.com/callback",
			},
			{
				name: "custom 301 RedirectResponse error",
				invoke: func(c *gin.Context) {
					handleResultWithParams(c, mo.Err[TestRes](&utils.RedirectResponse{
						Code:     http.StatusMovedPermanently,
						Location: "/moved",
					}))
				},
				wantCode:     http.StatusMovedPermanently,
				wantLocation: "/moved",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest("GET", "/test", nil)

				tc.invoke(c)

				if w.Code != tc.wantCode {
					t.Fatalf("got status %v, want %v", w.Code, tc.wantCode)
				}
				if got := w.Header().Get("Location"); got != tc.wantLocation {
					t.Fatalf("got Location %q, want %q", got, tc.wantLocation)
				}
			})
		}
	})
}

// TestHandleExtractError tests handleExtractError function
func TestHandleExtractError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handleExtractError(c, &JSON[TestReq]{}, errors.New("parse error"))

	if w.Code != http.StatusOK {
		t.Errorf("got status code %v, want %v", w.Code, http.StatusOK)
	}

	if !c.IsAborted() {
		t.Error("expected context to be aborted")
	}

	var resp utils.JsonResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal json response: %v", err)
	}
	if resp.Code != utils.CodeInvalidParameter {
		t.Errorf("got code %v, want %v", resp.Code, utils.CodeInvalidParameter)
	}

	if !strings.Contains(w.Body.String(), "request parse error") {
		t.Errorf("response body %v does not contain 'request parse error'", w.Body.String())
	}

	if !strings.Contains(w.Body.String(), "parse error") {
		t.Errorf("response body %v does not contain 'parse error'", w.Body.String())
	}
}
