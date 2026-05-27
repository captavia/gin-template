package di

import (
	"template/config"

	"github.com/nats-io/nats.go"
	"github.com/samber/do"
)

func ProvideNats(i *do.Injector, cfg *config.Config) {
	do.Provide(i, func(injector *do.Injector) (*nats.Conn, error) {
		return nats.Connect(cfg.Nats.URL)
	})
}
