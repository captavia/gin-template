package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/pelletier/go-toml/v2"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

type Config struct {
	App      AppConfig      `toml:"app"`
	Redis    RedisConfig    `toml:"redis"`
	Database DatabaseConfig `toml:"database"`
}

type AppConfig struct {
	Host string `toml:"host"`
}

type RedisConfig struct {
	Addr     string `toml:"addr"`
	Password string `toml:"password"`
	DB       int    `toml:"db"`
}

type DatabaseConfig struct {
	DSN string `toml:"dsn"`
}

func ifNilOr[T any](v T, or T) T {
	return lo.Ternary(lo.IsNil(v), v, or)
}

func DefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Host: ifNilOr(os.Getenv("APP_HOST"), "localhost:8080"),
		},
		Redis: RedisConfig{
			Addr:     ifNilOr(os.Getenv("REDIS_HOST"), "localhost:6379"),
			Password: ifNilOr(os.Getenv("REDIS_PASSWORD"), ""),
			DB:       int(ifNilOr(mo.TupleToResult(strconv.ParseInt(os.Getenv("REDIS_DB"), 10, 64)).OrElse(0), 0)),
		},
		Database: DatabaseConfig{
			DSN: ifNilOr(os.Getenv("DATABASE_DSN"), "postgres://postgres:P@ssword@localhost:5432/postgres?sslmode=disable&TimeZone=Asia/Shanghai"),
		},
	}
}

func LoadConfig(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// 如果配置文件不存在，生成默认配置
			defaultCfg := DefaultConfig()
			out, marshalErr := toml.Marshal(defaultCfg)
			if marshalErr != nil {
				return nil, marshalErr
			}
			if writeErr := os.WriteFile(path, out, 0644); writeErr != nil {
				return nil, writeErr
			}
			return defaultCfg, nil
		}
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
