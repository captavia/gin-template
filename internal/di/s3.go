package di

import (
	"errors"
	"template/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/samber/do/v2"
)

func ProvideS3(i do.Injector, cfg *config.Config) {
	do.Provide(i, func(injector do.Injector) (*minio.Client, error) {
		if cfg.S3.Endpoint == "" {
			return nil, errors.New("s3: endpoint is required")
		}
		if cfg.S3.AccessKeyID == "" {
			return nil, errors.New("s3: access key id is required")
		}
		if cfg.S3.SecretAccessKey == "" {
			return nil, errors.New("s3: secret access key is required")
		}

		return minio.New(cfg.S3.Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.S3.AccessKeyID, cfg.S3.SecretAccessKey, ""),
			Secure: cfg.S3.SSL,
			Region: cfg.S3.Region,
		})
	})
}
