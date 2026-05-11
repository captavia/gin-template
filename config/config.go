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
	Casbin   CasbinConfig   `toml:"casbin"`
	S3       S3Config       `toml:"s3"`
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
	DBType string `toml:"db_type"`
	DSN    string `toml:"dsn"`
	Prefix string `toml:"prefix"`
}

type CasbinConfig struct {
	ModelPath string `toml:"model_path"`
}

type S3Config struct {
	Endpoint        string `toml:"endpoint"`
	AccessKeyID     string `toml:"access_key_id"`
	SecretAccessKey string `toml:"secret_access_key"`
	Bucket          string `toml:"bucket"`
	Region          string `toml:"region"`
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
			DB:       int(mo.TupleToResult(strconv.ParseInt(ifNilOr(os.Getenv("REDIS_DB"), "0"), 10, 64)).OrElse(0)),
		},
		Database: DatabaseConfig{
			DBType: ifNilOr(os.Getenv("DATABASE_DB_TYPE"), "postgres"),
			DSN:    ifNilOr(os.Getenv("DATABASE_DSN"), "postgres://postgres:P@ssword@localhost:5432/postgres?sslmode=disable&TimeZone=Asia/Shanghai"),
			Prefix: ifNilOr(os.Getenv("DATABASE_PREFIX"), "template_"),
		},
		Casbin: CasbinConfig{
			ModelPath: ifNilOr(os.Getenv("CASBIN_MODEL_PATH"), "config/rbac_model.conf"),
		},
		S3: S3Config{
			Endpoint:        ifNilOr(os.Getenv("S3_ENDPOINT"), ""),
			AccessKeyID:     ifNilOr(os.Getenv("S3_ACCESS_KEY_ID"), ""),
			SecretAccessKey: ifNilOr(os.Getenv("S3_SECRET_ACCESS_KEY"), ""),
			Bucket:          ifNilOr(os.Getenv("S3_BUCKET"), ""),
			Region:          ifNilOr(os.Getenv("S3_REGION"), "us-east-1"),
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
