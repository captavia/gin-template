# Gin Template 项目开发指南

本文档为 AI Agent 和开发者提供项目架构、约定和工作流程的核心上下文。

## 项目概览

这是一个生产就绪的 Gin Web API 模板，主要特性：
- **类型安全的请求处理**：通过自定义 `pkg/wrap` 包实现泛型化请求绑定
- **依赖注入**：使用 `samber/do` 管理组件生命周期
- **Result 错误处理**：基于 `samber/mo` 的函数式错误处理
- **基础设施集成**：PostgreSQL/MySQL、Redis、MinIO S3、NATS
- **认证授权**：JWT + RBAC（基于 gorbac 持久化）

## 架构原则

### 1. 依赖注入模式
所有组件通过 `internal/di/container.go` 统一装配：
```go
// 基础设施提供者
ProvideZap, ProvideRedis, ProvideDB, ProvideS3, ProvideNats

// 服务层
service.NewJwtService, service.NewRBACService, service.NewAuthService

// 处理器层
handler.NewAuthHandler
```

**规则**：永远不要手动实例化依赖，始终使用 `do.MustInvoke[T](injector)`。

### 2. 类型安全的请求处理
项目使用自定义的 `pkg/wrap` 包来消除样板代码：

#### WrapTyped（遗留方式）
```go
type authRequest struct {
    Phone    string `json:"phone" binding:"required"`
    Password string `json:"password" binding:"required,min=6"`
}

func (h *AuthHandler) Register(c *gin.Context, req *authRequest) mo.Result[string] {
    // 处理逻辑
}

// 路由注册
apiV1.POST("/auth/register", wrap.WrapTyped(authHandler.Register))
```

#### Wrap1-5（推荐方式）
使用 `Extractor` 接口实现灵活的参数提取：
```go
type Extractor interface {
    Extract(c *gin.Context) error
}
```

内置提取器：
- `wrap.JSON[T]` - 解析 JSON 请求体
- `wrap.Query[T]` - 解析查询参数
- `wrap.Path[T]` - 解析 URL 路径参数
- `wrap.File[T]` - 解析 multipart 表单数据

示例：
```go
func (h *AuthHandler) Login(c *gin.Context, req *wrap.JSON[authRequest]) mo.Result[LoginResponse] {
    // 通过 req.Data 访问数据
    return mo.Ok(LoginResponse{Token: req.Data.Phone})
}

// 路由注册
apiV1.POST("/auth/login", wrap.Wrap1(authHandler.Login))
```

**多个参数**（使用 Wrap2、Wrap3 等）：
```go
type UserIDPath struct {
    ID uint `uri:"id" binding:"required"`
}

type UpdateUserReq struct {
    Name string `json:"name"`
}

func (h *UserHandler) UpdateUser(
    c *gin.Context,
    path *wrap.Path[UserIDPath],
    body *wrap.JSON[UpdateUserReq],
) mo.Result[string] {
    userID := path.Data.ID
    newName := body.Data.Name
    // ...
}

// 路由注册
apiV1.PUT("/users/:id", wrap.Wrap2(userHandler.UpdateUser))
```

#### 高级用法：自定义验证
实现 `Defaulter` 或 `Validator` 接口：
```go
type CreatePostReq struct {
    Title   string `json:"title"`
    Content string `json:"content"`
    Status  string `json:"status"`
}

func (r *CreatePostReq) Default() {
    if r.Status == "" {
        r.Status = "draft"
    }
}

func (r *CreatePostReq) Validate() error {
    if len(r.Title) > 200 {
        return errors.New("标题过长")
    }
    return nil
}
```

Extractor 会在绑定前自动调用 `Default()`，在绑定后自动调用 `Validate()`。

### 3. Result 错误处理
所有 Handler 返回值使用 `mo.Result[T]`：
```go
// 成功
return mo.Ok(data)

// 业务错误（带自定义状态码）
return mo.Err[T](utils.Err(http.StatusConflict, errors.New("手机号已存在")))

// 通用错误（500）
return mo.Err[T](err)
```

响应结构（由 wrap 自动处理）：
```json
{
  "code": 200,
  "data": {...},
  "message": "success"
}
```

错误响应：
```json
{
  "code": 409,
  "error": "手机号已存在"
}
```

## 目录结构

```
├── cmd/api/              # 应用程序入口
│   ├── main.go          # 加载配置、构建 DI 容器、启动服务器
│   └── config.toml      # 运行时配置（已加入 .gitignore）
│
├── config/              # 配置管理
│   └── config.go        # TOML + 环境变量支持，带默认值
│
├── internal/
│   ├── api/
│   │   ├── handler/     # HTTP 请求处理器（薄层）
│   │   ├── router/      # 路由定义
│   │   ├── service/     # 业务逻辑层
│   │   └── query/       # 数据库查询辅助
│   ├── di/              # 依赖注入
│   │   ├── container.go # DI 容器构建器
│   │   ├── db.go        # 数据库提供者
│   │   ├── redis.go     # Redis 提供者
│   │   ├── s3.go        # MinIO S3 提供者
│   │   ├── nats.go      # NATS 提供者
│   │   └── zap.go       # 日志提供者
│   ├── middleware/      # Gin 中间件（认证、日志等）
│   └── model/           # GORM 模型
│
└── pkg/                 # 可复用库
    ├── wrap/            # 类型安全请求包装器
    │   ├── core.go      # WrapTyped、Wrap、Wrap1-5
    │   └── extractor.go # Extractor 接口及实现
    └── utils/           # 工具函数
        └── response.go  # 响应辅助函数（Ok、Err）
```

## 开发工作流

### 添加新的 API 端点

1. **定义请求/响应类型**：
```go
// internal/api/handler/post.go
type CreatePostReq struct {
    Title   string `json:"title" binding:"required"`
    Content string `json:"content" binding:"required"`
}

type PostResponse struct {
    ID        uint      `json:"id"`
    Title     string    `json:"title"`
    CreatedAt time.Time `json:"created_at"`
}
```

2. **创建 Handler**（始终通过 DI 注入依赖）：
```go
type PostHandler struct {
    postService *service.PostService
}

func NewPostHandler(i do.Injector) (*PostHandler, error) {
    return &PostHandler{
        postService: do.MustInvoke[*service.PostService](i),
    }, nil
}

func (h *PostHandler) CreatePost(
    c *gin.Context,
    req *wrap.JSON[CreatePostReq],
) mo.Result[PostResponse] {
    post, err := h.postService.CreatePost(c.Request.Context(), req.Data.Title, req.Data.Content)
    if err != nil {
        return mo.Err[PostResponse](utils.Err(http.StatusInternalServerError, err))
    }
    return mo.Ok(PostResponse{
        ID:        post.ID,
        Title:     post.Title,
        CreatedAt: post.CreatedAt,
    })
}
```

3. **实现 Service**：
```go
// internal/api/service/post.go
type PostService struct {
    db *gorm.DB
}

func NewPostService(i do.Injector) (*PostService, error) {
    return &PostService{
        db: do.MustInvoke[*gorm.DB](i),
    }, nil
}

func (s *PostService) CreatePost(ctx context.Context, title, content string) (*model.Post, error) {
    post := &model.Post{Title: title, Content: content}
    if err := s.db.WithContext(ctx).Create(post).Error; err != nil {
        return nil, err
    }
    return post, nil
}
```

4. **在 DI 容器中注册**：
```go
// internal/di/container.go
func BuildContainer(cfg *config.Config) do.Injector {
    // ...
    
    // 添加到 Services 层
    do.Provide(injector, service.NewPostService)
    
    // 添加到 Handlers 层
    do.Provide(injector, handler.NewPostHandler)
    
    return injector
}
```

5. **添加路由**：
```go
// internal/api/router/router.go
func SetupRouter(injector do.Injector) *gin.Engine {
    // ...
    postHandler := do.MustInvoke[*handler.PostHandler](injector)
    
    apiV1.POST("/posts", wrap.Wrap1(postHandler.CreatePost))
    return r
}
```

### 添加新的 Model

```go
// internal/model/post.go
package model

import "time"

type Post struct {
    ID        uint      `gorm:"primaryKey"`
    Title     string    `gorm:"type:varchar(200);not null"`
    Content   string    `gorm:"type:text"`
    UserID    uint      `gorm:"index;not null"`
    User      User      `gorm:"foreignKey:UserID"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### 添加新的依赖提供者

```go
// internal/di/new_service.go
package di

import (
    "template/config"
    "github.com/samber/do/v2"
)

func ProvideNewService(i do.Injector, cfg *config.Config) {
    do.ProvideNamed(i, "new_service", func(i do.Injector) (*NewService, error) {
        // 使用 cfg 初始化服务
        return newService, nil
    })
}

// 在 BuildContainer 中调用：
// ProvideNewService(injector, cfg)
```

## 代码约定

### 命名规范
- **Handler**：`{资源}Handler`（例如：`AuthHandler`、`UserHandler`）
- **Service**：`{资源}Service`（例如：`UserService`、`PostService`）
- **请求结构体**：`{操作}{资源}Req`（例如：`CreatePostReq`、`UpdateUserReq`）
- **响应结构体**：`{资源}Response`（例如：`PostResponse`、`LoginResponse`）

### Handler 职责
Handler 应该是薄适配层：
- 提取请求参数（通过 wrap）
- 调用 Service 层
- 将 Service 结果映射为 HTTP 响应
- **绝不**包含业务逻辑

### Service 层
- 包含所有业务逻辑
- 验证业务规则
- 编排数据库操作
- 返回领域错误（而非 HTTP 错误）

### 错误处理
- **业务错误**：使用 `utils.Err(code, err)` 处理客户端错误（400-499）
- **系统错误**：返回原始 `error` 表示服务器错误（500）
- 始终使用 `context.Context` 支持取消操作

### 数据库操作
- 始终使用 `WithContext(ctx)` 支持取消
- 多步操作使用事务：
```go
return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    // 多步操作
    return nil
})
```

## 配置管理

配置优先级：**环境变量 > TOML 文件 > 默认值**

添加新配置字段：
```go
// config/config.go
type Config struct {
    // ...
    NewService NewServiceConfig `toml:"new_service"`
}

type NewServiceConfig struct {
    APIKey string `toml:"api_key"`
}

// 更新 DefaultConfig()：
NewService: NewServiceConfig{
    APIKey: ifNilOr(os.Getenv("NEW_SERVICE_API_KEY"), ""),
}
```

## 测试

项目使用 Go 标准测试包。测试结构示例：

```go
// pkg/wrap/core_test.go
func TestWrap1(t *testing.T) {
    // 设置 Gin 测试上下文
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    
    // 测试逻辑
}
```

运行测试：
```bash
go test ./...
go test -v ./pkg/wrap/  # 特定包
go test -cover ./...     # 带覆盖率
```

## 构建和运行

```bash
# 构建优化后的二进制文件
make all

# 直接运行
go run ./cmd/api

# 运行二进制文件
./build/api
```

构建标志说明：
- `-ldflags="-s -w"`：去除调试信息（更小的二进制文件）
- `-trimpath`：从二进制文件中移除文件路径
- `-tags=nomsgpack`：排除 msgpack（更轻量的构建）

## 常用模式

### 需要认证的端点
```go
// 使用中间件（在 internal/middleware/auth.go 实现）
authorized := apiV1.Group("/")
authorized.Use(middleware.RequireAuth())
{
    authorized.GET("/profile", wrap.Wrap(userHandler.GetProfile))
}
```

### 分页
```go
type PaginationQuery struct {
    Page     int `form:"page" binding:"min=1"`
    PageSize int `form:"page_size" binding:"min=1,max=100"`
}

func (p *PaginationQuery) Default() {
    if p.Page == 0 {
        p.Page = 1
    }
    if p.PageSize == 0 {
        p.PageSize = 20
    }
}

// Handler
func (h *PostHandler) ListPosts(
    c *gin.Context,
    query *wrap.Query[PaginationQuery],
) mo.Result[[]PostResponse] {
    posts := h.postService.ListPosts(query.Data.Page, query.Data.PageSize)
    return mo.Ok(posts)
}
```

### 文件上传
```go
type UploadFileReq struct {
    File *multipart.FileHeader `form:"file" binding:"required"`
}

func (h *FileHandler) Upload(
    c *gin.Context,
    req *wrap.File[UploadFileReq],
) mo.Result[string] {
    file := req.Data.File
    // 处理文件
}
```

## 核心依赖

- **gin-gonic/gin**：Web 框架
- **samber/do**：依赖注入
- **samber/mo**：Result 类型（Option/Result monad）
- **gorm.io/gorm**：ORM
- **redis/go-redis**：Redis 客户端
- **minio/minio-go**：S3 兼容存储
- **nats-io/nats.go**：消息队列
- **golang-jwt/jwt**：JWT 令牌
- **mikespook/gorbac**：RBAC 实现

## 故障排查

### "invalid parameter" 错误
检查：
1. 请求结构体标签与 HTTP 方法匹配（`json`、`form`、`uri`）
2. `binding` 标签正确
3. Extractor 类型与数据源匹配（JSON/Query/Path/File）

### DI 错误（"service not found"）
确保服务已在 `internal/di/container.go` 中注册：
```go
do.Provide(injector, service.NewYourService)
```

### 配置未加载
1. 检查 `cmd/api/config.toml` 存在（首次运行自动生成）
2. 验证环境变量已设置
3. 检查字段名与 TOML 键匹配（区分大小写）

## 性能注意事项

- **Wrap vs Wrap1-5**：Wrap1-5 性能更高（无反射）但类型受限
- **Wrap（反射）**：用于 >5 个参数或参数数量可变的情况
- GORM、Redis 客户端和 NATS 自动处理连接池
- 使用预编译语句（GORM 自动处理）

## 安全检查清单

- [ ] 永远不要记录敏感数据（密码、令牌）
- [ ] 验证所有用户输入（使用 `binding` 标签 + 自定义验证器）
- [ ] 使用参数化查询（GORM 默认）
- [ ] 为认证端点实现速率限制
- [ ] 生产环境使用 HTTPS
- [ ] 定期轮换 JWT 密钥
- [ ] 设置适当的令牌过期时间
