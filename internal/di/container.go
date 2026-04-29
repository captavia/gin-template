package di

import (
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/samber/do"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"template/config"
	"template/internal/api/handler"
	"template/internal/model"
	"template/internal/service"
)

func BuildContainer(cfg *config.Config) *do.Injector {
	injector := do.New()

	do.ProvideValue(injector, cfg)

	do.ProvideValue(injector, redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}))

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "betting_",
		},
	})
	if err != nil {
		log.Println("Warning: Database connection failed (Mocking DB locally maybe?):", err.Error())
	} else {
		// 启动时自动迁移表结构
		err = db.AutoMigrate(
			&model.User{},
		)
		if err != nil {
			log.Println("AutoMigrate failed:", err.Error())
		}
	}
	do.ProvideValue[*gorm.DB](injector, db)

	// 3. 注册 Services 层
	// 使用依赖注入装载服务实例 (由框架自定分析并满足它们需要的 *redis.Client 等)
	do.Provide(injector, service.NewAuthService)

	// 4. 注册 Handlers层
	// 组装的时候会自动注入以上准备好的 Service
	do.Provide(injector, handler.NewAuthHandler)

	return injector
}
