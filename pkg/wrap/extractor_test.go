package wrap

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// Test structures
type PathData struct {
	ID string `uri:"id" binding:"required"`
}

type QueryData struct {
	Name  string `form:"name"`
	Age   int    `form:"age"`
	Email string `form:"email"`
}

type JSONData struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type FileData struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
	Name string                `form:"name"`
}

type DefaultableData struct {
	Value   string
	Applied bool
}

func (d *DefaultableData) Default() {
	d.Value = "default_value"
	d.Applied = true
}

type ValidatableData struct {
	Value string
}

func (v *ValidatableData) Validate() error {
	if v.Value == "invalid" {
		return errors.New("validation failed")
	}
	return nil
}

// TestPathExtract tests Path extractor
func TestPathExtract(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		params    map[string]string
		wantData  PathData
		wantError bool
	}{
		{
			name:      "valid path param",
			path:      "/users/:id",
			params:    map[string]string{"id": "123"},
			wantData:  PathData{ID: "123"},
			wantError: false,
		},
		{
			name:      "missing required param",
			path:      "/users/:id",
			params:    map[string]string{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", tt.path, nil)

			// Set URL params
			c.Params = []gin.Param{}
			for k, v := range tt.params {
				c.Params = append(c.Params, gin.Param{Key: k, Value: v})
			}

			p := &Path[PathData]{}
			err := p.Extract(c)

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if p.Data.ID != tt.wantData.ID {
				t.Errorf("got ID %v, want %v", p.Data.ID, tt.wantData.ID)
			}
		})
	}
}

// TestQueryExtract tests Query extractor
func TestQueryExtract(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantData  QueryData
		wantError bool
	}{
		{
			name:  "valid query params",
			query: "name=John&age=30&email=john@example.com",
			wantData: QueryData{
				Name:  "John",
				Age:   30,
				Email: "john@example.com",
			},
			wantError: false,
		},
		{
			name:  "partial query params",
			query: "name=Jane",
			wantData: QueryData{
				Name: "Jane",
				Age:  0,
			},
			wantError: false,
		},
		{
			name:      "empty query",
			query:     "",
			wantData:  QueryData{},
			wantError: false,
		},
		{
			name:      "invalid type conversion",
			query:     "age=not_a_number",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test?"+tt.query, nil)

			q := &Query[QueryData]{}
			err := q.Extract(c)

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if q.Data.Name != tt.wantData.Name {
				t.Errorf("got Name %v, want %v", q.Data.Name, tt.wantData.Name)
			}
			if q.Data.Age != tt.wantData.Age {
				t.Errorf("got Age %v, want %v", q.Data.Age, tt.wantData.Age)
			}
		})
	}
}

// TestJSONExtract tests JSON extractor
func TestJSONExtract(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantData  JSONData
		wantError bool
	}{
		{
			name:      "valid JSON",
			body:      `{"username":"testuser","password":"testpass"}`,
			wantData:  JSONData{Username: "testuser", Password: "testpass"},
			wantError: false,
		},
		{
			name:      "missing required field",
			body:      `{"username":"testuser"}`,
			wantError: true,
		},
		{
			name:      "invalid JSON",
			body:      `{invalid json}`,
			wantError: true,
		},
		{
			name:      "empty body",
			body:      `{}`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/test", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			j := &JSON[JSONData]{}
			err := j.Extract(c)

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if j.Data.Username != tt.wantData.Username {
				t.Errorf("got Username %v, want %v", j.Data.Username, tt.wantData.Username)
			}
			if j.Data.Password != tt.wantData.Password {
				t.Errorf("got Password %v, want %v", j.Data.Password, tt.wantData.Password)
			}
		})
	}
}

// TestFileExtract tests File extractor
func TestFileExtract(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (*bytes.Buffer, string)
		wantName  string
		wantError bool
	}{
		{
			name: "valid file upload",
			setup: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				part, _ := writer.CreateFormFile("file", "test.txt")
				part.Write([]byte("test content"))
				writer.WriteField("name", "test_file")
				writer.Close()

				return body, writer.FormDataContentType()
			},
			wantName:  "test_file",
			wantError: false,
		},
		{
			name: "missing file",
			setup: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.WriteField("name", "no_file")
				writer.Close()

				return body, writer.FormDataContentType()
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, contentType := tt.setup()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/test", body)
			c.Request.Header.Set("Content-Type", contentType)

			f := &File[FileData]{}
			err := f.Extract(c)

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if f.Data.Name != tt.wantName {
				t.Errorf("got Name %v, want %v", f.Data.Name, tt.wantName)
			}
			if f.Data.File == nil {
				t.Error("expected file but got nil")
			}
		})
	}
}

// TestDefaulter tests the Defaulter interface
func TestDefaulter(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	j := &JSON[DefaultableData]{}
	err := j.Extract(c)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if !j.Data.Applied {
		t.Error("Default() was not called")
	}
	if j.Data.Value != "default_value" {
		t.Errorf("got Value %v, want default_value", j.Data.Value)
	}
}

// TestValidator tests the Validator interface
func TestValidator(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{
			name:      "valid value",
			value:     "valid",
			wantError: false,
		},
		{
			name:      "invalid value",
			value:     "invalid",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			body := fmt.Sprintf(`{"Value":"%s"}`, tt.value)
			c.Request = httptest.NewRequest("POST", "/test", strings.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")

			j := &JSON[ValidatableData]{}
			err := j.Extract(c)

			if tt.wantError {
				if err == nil {
					t.Error("expected validation error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestSmartString tests smartString function
func TestSmartString(t *testing.T) {
	// Save original value
	originalMax := maxOutputLen.Load()
	defer maxOutputLen.Store(originalMax)

	tests := []struct {
		name     string
		maxLen   int32
		data     any
		contains string
	}{
		{
			name:     "short string",
			maxLen:   100,
			data:     "hello",
			contains: "hello",
		},
		{
			name:     "long struct truncated",
			maxLen:   10,
			data:     struct{ A, B, C, D, E string }{"a", "b", "c", "d", "e"},
			contains: "output too long",
		},
		{
			name:     "nil pointer",
			maxLen:   100,
			data:     (*string)(nil),
			contains: "nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maxOutputLen.Store(tt.maxLen)
			result := smartString(tt.data)

			if !strings.Contains(result, tt.contains) {
				t.Errorf("smartString() = %v, want to contain %v", result, tt.contains)
			}
		})
	}
}

// TestSetMaxOutputLen tests SetMaxOutputLen function
func TestSetMaxOutputLen(t *testing.T) {
	originalMax := maxOutputLen.Load()
	defer maxOutputLen.Store(originalMax)

	SetMaxOutputLen(1000)
	if maxOutputLen.Load() != 1000 {
		t.Errorf("got %v, want 1000", maxOutputLen.Load())
	}

	SetMaxOutputLen(100)
	if maxOutputLen.Load() != 100 {
		t.Errorf("got %v, want 100", maxOutputLen.Load())
	}
}

// TestExtractorString tests String() methods for all extractors
func TestExtractorString(t *testing.T) {
	tests := []struct {
		name      string
		extractor fmt.Stringer
		contains  string
	}{
		{
			name:      "Path String",
			extractor: &Path[PathData]{Data: PathData{ID: "123"}},
			contains:  "123",
		},
		{
			name:      "Query String",
			extractor: &Query[QueryData]{Data: QueryData{Name: "John", Age: 30}},
			contains:  "John",
		},
		{
			name:      "JSON String",
			extractor: &JSON[JSONData]{Data: JSONData{Username: "test"}},
			contains:  "test",
		},
		{
			name:      "File String",
			extractor: &File[FileData]{Data: FileData{Name: "test.txt"}},
			contains:  "test.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.extractor.String()
			if !strings.Contains(result, tt.contains) {
				t.Errorf("String() = %v, want to contain %v", result, tt.contains)
			}
		})
	}
}
