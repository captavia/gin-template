package provider

import (
	"gin-template/pkg/config"
	"github.com/samber/do"
)

func Config(path string) func(*do.Injector) (*config.Config, error) {
	return func(injector *do.Injector) (*config.Config, error) {
		return config.LoadConfig(path)
	}
}
