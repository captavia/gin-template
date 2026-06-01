package di

import (
	"template/config"
	"template/internal/api/handler"
	"template/internal/service"

	"github.com/samber/do/v2"
)

func BuildContainer(cfg *config.Config) do.Injector {
	injector := do.New()

	do.ProvideValue(injector, cfg)

	// 注册日志
	ProvideZap(injector, cfg)

	// 注册redis
	ProvideRedis(injector, cfg)

	// 注册数据库
	ProvideDB(injector, cfg)

	// 注册 S3
	ProvideS3(injector, cfg)

	// 注册 NATS
	ProvideNats(injector, cfg)

	// 3. 注册 Services 层
	// 使用依赖注入装载服务实例 (由框架自动分析并满足它们需要的依赖)
	do.Provide(injector, service.NewJwtService)
	do.Provide(injector, service.NewAuthService)
	do.Provide(injector, service.NewRBACService)

	// 4. 注册 Handlers层
	// 组装的时候会自动注入以上准备好的 Service
	do.Provide(injector, handler.NewAuthHandler)

	return injector
}
