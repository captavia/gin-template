package di

import (
	"template/config"

	"github.com/redis/go-redis/v9"
	"github.com/samber/do"
)

func ProvideRedis(i *do.Injector, cfg *config.Config) {
	do.ProvideValue(i, redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}))
}
