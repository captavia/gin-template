package wrap

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/samber/mo"
)

// 测试用的请求结构
type BenchRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Age     int    `json:"age"`
	Address string `json:"address"`
}

// 小结构体（32 bytes）
type SmallRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// 大结构体（~500 bytes）
type LargeRequest struct {
	Field1  string  `json:"field1"`
	Field2  string  `json:"field2"`
	Field3  string  `json:"field3"`
	Field4  string  `json:"field4"`
	Field5  string  `json:"field5"`
	Field6  string  `json:"field6"`
	Field7  string  `json:"field7"`
	Field8  string  `json:"field8"`
	Field9  string  `json:"field9"`
	Field10 string  `json:"field10"`
	Numbers [50]int `json:"numbers"`
}

type BenchResponse struct {
	Status string `json:"status"`
}

// 值传递版本的 handler
func handlerValue(req JSON[BenchRequest]) mo.Result[BenchResponse] {
	return mo.Ok(BenchResponse{Status: "ok"})
}

// 指针传递版本的 handler（模拟旧实现）
func handlerPointer(req *JSON[BenchRequest]) mo.Result[BenchResponse] {
	return mo.Ok(BenchResponse{Status: "ok"})
}

func handlerSmallValue(req JSON[SmallRequest]) mo.Result[BenchResponse] {
	return mo.Ok(BenchResponse{Status: "ok"})
}

func handlerLargeValue(req JSON[LargeRequest]) mo.Result[BenchResponse] {
	return mo.Ok(BenchResponse{Status: "ok"})
}

// 模拟旧的指针版本 Wrap1
func Wrap1Pointer[T1 any, PT1 interface {
	*T1
	Extractor
}, Res any](h func(t1 PT1) mo.Result[Res]) gin.HandlerFunc {
	return func(c *gin.Context) {
		t1 := new(T1)
		if err := PT1(t1).Extract(c); err != nil {
			handleExtractError(c, err)
			return
		}
		handleResult(c, h(PT1(t1)))
	}
}

func BenchmarkWrap1_Value_Normal(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	handler := Wrap1(handlerValue)

	jsonBody := `{"name":"test","email":"test@example.com","age":25,"address":"123 Main St"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler(c)
	}
}

func BenchmarkWrap1_Pointer_Normal(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	handler := Wrap1Pointer(handlerPointer)

	jsonBody := `{"name":"test","email":"test@example.com","age":25,"address":"123 Main St"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler(c)
	}
}

func BenchmarkWrap1_Value_Small(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	handler := Wrap1(handlerSmallValue)

	jsonBody := `{"id":"123","name":"test"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler(c)
	}
}

func BenchmarkWrap1_Value_Large(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	handler := Wrap1(handlerLargeValue)

	jsonBody := `{"field1":"a","field2":"b","field3":"c","field4":"d","field5":"e","field6":"f","field7":"g","field8":"h","field9":"i","field10":"j","numbers":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50]}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler(c)
	}
}

// 测试纯值拷贝的开销（不包含网络和 JSON 解析）
func BenchmarkPureCopy_Normal(b *testing.B) {
	data := JSON[BenchRequest]{
		Data: BenchRequest{
			Name:    "test",
			Email:   "test@example.com",
			Age:     25,
			Address: "123 Main St",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = copyByValue(data)
	}
}

func BenchmarkPureCopy_Large(b *testing.B) {
	data := JSON[LargeRequest]{
		Data: LargeRequest{
			Field1: "a", Field2: "b", Field3: "c",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = copyByValueLarge(data)
	}
}

func copyByValue(data JSON[BenchRequest]) JSON[BenchRequest] {
	return data
}

func copyByValueLarge(data JSON[LargeRequest]) JSON[LargeRequest] {
	return data
}
