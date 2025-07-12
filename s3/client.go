package s3

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log/slog"
	"os"
)

// S3SelectelClient selectel implementation
func NewSelectelS3Client(log *slog.Logger) (*s3.Client, error) {
	endpoint := os.Getenv("AWS_ENDPOINT")
	//accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	//secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_REGION")
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithEndpointResolver(aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			if service == s3.ServiceID {
				return aws.Endpoint{
					URL:               endpoint,
					HostnameImmutable: true,
					SigningRegion:     region,
				}, nil
			}
			return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
		})),
	)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = false
	})
	return client, nil
}
