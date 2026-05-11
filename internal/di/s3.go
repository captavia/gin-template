package di

import (
	"context"
	"log"
	"template/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/samber/do"
)

func ProvideS3(i *do.Injector, cfg *config.Config) {
	do.ProvideValue(i, func() *s3.Client {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if cfg.S3.Endpoint != "" {
				return aws.Endpoint{
					PartitionID:       "aws",
					URL:               cfg.S3.Endpoint,
					SigningRegion:     cfg.S3.Region,
					HostnameImmutable: true,
				}, nil
			}
			// returning EndpointNotFoundError will allow the service to fall back to its default resolution
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		})

		awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
			awsconfig.WithRegion(cfg.S3.Region),
			awsconfig.WithEndpointResolverWithOptions(customResolver),
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.S3.AccessKeyID,
				cfg.S3.SecretAccessKey,
				"",
			)),
		)
		if err != nil {
			log.Fatalf("[Fatal] Failed to load AWS S3 config: %v", err)
		}

		client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			if cfg.S3.Endpoint != "" {
				o.UsePathStyle = true
			}
		})
		return client
	}())

}
