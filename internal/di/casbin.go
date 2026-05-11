package di

import (
	"template/config"

	"github.com/casbin/casbin/v3"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func ProvideCasbin(i *do.Injector, cfg *config.Config) {
	do.Provide(i, func(i *do.Injector) (*casbin.Enforcer, error) {
		db, err := do.Invoke[*gorm.DB](i)
		if err != nil {
			return nil, err
		}

		adapter, err := gormadapter.NewAdapterByDB(db)
		if err != nil {
			return nil, err
		}

		enforcer, err := casbin.NewEnforcer(cfg.Casbin.ModelPath, adapter)
		if err != nil {
			return nil, err
		}

		// 开启自动保存，这样在内存中修改策略后会自动同步到数据库
		enforcer.EnableAutoSave(true)

		// 加载策略
		if err := enforcer.LoadPolicy(); err != nil {
			return nil, err
		}

		return enforcer, nil
	})
}
