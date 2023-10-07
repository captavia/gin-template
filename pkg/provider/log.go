package provider

import (
	"gin-template/pkg/log"
	"github.com/samber/do"
	"github.com/sirupsen/logrus"
)

func Formatter() func(*do.Injector) (logrus.Formatter, error) {
	return func(injector *do.Injector) (logrus.Formatter, error) {
		return log.NewFormatter(), nil
	}
}
