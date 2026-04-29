package di

import (
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/samber/do"
	"github.com/samber/mo"
	"gorm.io/driver/mysql"
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

	do.ProvideValue[*gorm.DB](injector, mo.Do(func() *gorm.DB {
		dialer := func(dbType, dsn string) mo.Result[gorm.Dialector] {
			switch dbType {
			case "mysql":
				return mo.Ok(mysql.Open(dsn))
			case "postgres":
				return mo.Ok(postgres.Open(dsn))
			default:
				return mo.Err[gorm.Dialector](fmt.Errorf("unsupported database type: %s", dbType))
			}
		}(cfg.Database.DBType, cfg.Database.DSN).
			MapErr(func(err error) (gorm.Dialector, error) {
				log.Printf("[Fatal] Database config error: %v\n", err)
				return nil, err
			}).MustGet()

		// 2. 建立数据库连接
		db := mo.TupleToResult(
			gorm.Open(dialer, &gorm.Config{
				NamingStrategy: schema.NamingStrategy{
					TablePrefix: cfg.Database.Prefix,
				},
			}),
		).MapErr(func(err error) (*gorm.DB, error) {
			log.Printf("[Fatal] Database connection failed: %v\n", err)
			return nil, err
		}).MustGet()

		// 3. 执行自动迁移
		_ = mo.TupleToResult(db, db.AutoMigrate(
			new(model.User),
		)).MapErr(
			func(err error) (*gorm.DB, error) {
				log.Printf("[Fatal] AutoMigrate failed: %v\n", err)
				return nil, err
			},
		).MustGet()

		return db
	}).MustGet())

	// 3. 注册 Services 层
	// 使用依赖注入装载服务实例 (由框架自定分析并满足它们需要的 *redis.Client 等)
	do.Provide(injector, service.NewAuthService)

	// 4. 注册 Handlers层
	// 组装的时候会自动注入以上准备好的 Service
	do.Provide(injector, handler.NewAuthHandler)

	return injector
}
