package s3

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

// S3Client provides interaction with selectel object storage (S3)
type S3Client struct {
	s3        *s3.Client
	s3PreSign *s3.PresignClient
	logger    *slog.Logger
}

func NewS3Client(log *slog.Logger) (*S3Client, error) {
	s3Client, err := NewSelectelS3Client(log)
	if err != nil {
		panic(err)
	}
	presigner := s3.NewPresignClient(s3Client, func(o *s3.PresignOptions) {
		o.ClientOptions = append(o.ClientOptions, func(opts *s3.Options) {
			opts.UsePathStyle = false
		})
	})

	return &S3Client{
		s3:        s3Client,
		s3PreSign: presigner,
		logger:    log,
	}, nil
}

// GeneratePreSignUrl generates pre-signed url from s3
func (s *S3Client) GeneratePreSignUrl(ctx context.Context, bucket, contentType string) (string, string, error) {
	key := uuid.New().String()
	timeToExpire := 1 * time.Hour
	keyWithType := fmt.Sprintf("%s.%s", key, contentType)
	preSignedRes, err := s.s3PreSign.PresignPutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(keyWithType),
		}, s3.WithPresignExpires(timeToExpire),
	)
	if err != nil {
		return "", "", fmt.Errorf("error generating pre-signed url: %v", err)
	}
	return preSignedRes.URL, key, nil
}
