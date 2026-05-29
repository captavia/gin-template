package di

import (
	"strings"

	"github.com/samber/do"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"template/config"
)

func ProvideZap(injector *do.Injector, cfg *config.Config) {
	do.Provide(injector, func(i *do.Injector) (*zap.Logger, error) {
		level := zapcore.InfoLevel
		switch strings.ToLower(cfg.Log.Level) {
		case "debug":
			level = zapcore.DebugLevel
		case "info":
			level = zapcore.InfoLevel
		case "warn":
			level = zapcore.WarnLevel
		case "error":
			level = zapcore.ErrorLevel
		}

		zapConfig := zap.NewProductionConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(level)

		return zapConfig.Build(zap.AddCaller())
	})
}
