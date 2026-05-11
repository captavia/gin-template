package di

import (
	"github.com/samber/do"

	"template/config"
	"template/internal/api/handler"
	"template/internal/service"
)

func BuildContainer(cfg *config.Config) *do.Injector {
	injector := do.New()

	do.ProvideValue(injector, cfg)

	// 注册redis
	ProvideRedis(injector, cfg)

	// 注册数据库
	ProvideDB(injector, cfg)

	// 注册 Casbin
	ProvideCasbin(injector, cfg)

	// 注册 S3
	ProvideS3(injector, cfg)

	// 3. 注册 Services 层
	// 使用依赖注入装载服务实例 (由框架自定分析并满足它们需要的 *redis.Client 等)
	do.Provide(injector, service.NewAuthService)
	do.Provide(injector, service.NewPermissionService)

	// 4. 注册 Handlers层
	// 组装的时候会自动注入以上准备好的 Service
	do.Provide(injector, handler.NewAuthHandler)
	do.Provide(injector, handler.NewPermissionHandler)

	return injector
}
