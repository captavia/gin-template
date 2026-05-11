package di

import (
	"fmt"
	"log"
	"template/config"
	"template/internal/model"

	"github.com/samber/do"
	"github.com/samber/mo"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func ProvideDB(i *do.Injector, cfg *config.Config) {
	do.ProvideValue[*gorm.DB](i, mo.Do(func() *gorm.DB {
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
}
