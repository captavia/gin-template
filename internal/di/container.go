package di

import (
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/samber/do"
	"github.com/samber/mo"
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

	do.ProvideValue[*gorm.DB](injector, mo.
		TupleToResult(
			gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{
				NamingStrategy: schema.NamingStrategy{
					TablePrefix: cfg.Database.Prefix,
				},
			}),
		).
		MapErr(func(err error) (*gorm.DB, error) {
			log.Panicln("Warning: Database connection failed:", err)
			return nil, err
		}).
		// 3. 如果连接成功，则继续执行 AutoMigrate
		FlatMap(func(db *gorm.DB) mo.Result[*gorm.DB] {
			return mo.TupleToResult(db, db.AutoMigrate(
				new(model.User),
			)).MapErr(
				func(err error) (*gorm.DB, error) {
					log.Panicln("AutoMigrate failed:", err)
					return db, err
				},
			)
		}).MustGet())

	// 3. 注册 Services 层
	// 使用依赖注入装载服务实例 (由框架自定分析并满足它们需要的 *redis.Client 等)
	do.Provide(injector, service.NewAuthService)

	// 4. 注册 Handlers层
	// 组装的时候会自动注入以上准备好的 Service
	do.Provide(injector, handler.NewAuthHandler)

	return injector
}
