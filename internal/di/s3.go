package di

import (
	"template/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/samber/do"
)

func ProvideS3(i *do.Injector, cfg *config.Config) {
	do.Provide(i, func(injector *do.Injector) (*minio.Client, error) {
		return minio.New(cfg.S3.Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.S3.AccessKeyID, cfg.S3.SecretAccessKey, ""),
			Secure: cfg.S3.SSL,
		})
	})
}
