package di

import (
	"template/config"
	"template/internal/api/handler"
	"template/internal/api/service"

	"github.com/samber/do/v2"
)

func BuildContainer(cfg *config.Config) do.Injector {
	injector := do.New()

	do.ProvideValue(injector, cfg)

	//注册依赖
	{
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
	}

	// 注册 Services 层
	{
		do.Provide(injector, service.NewJwtService)
		do.Provide(injector, service.NewAuthService)
		do.Provide(injector, service.NewRBACService)
	}

	// 注册 Handlers层
	{
		do.Provide(injector, handler.NewAuthHandler)
	}

	return injector
}
