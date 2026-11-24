package storage

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"

	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/config"
)

var ErrObjectNotFound = errors.New("objeto não encontrado no storage")

type ObjectMetadata struct {
	SizeBytes int64
	MimeType  string
}

type R2Storage struct {
	client    *s3.Client
	presigner *s3.PresignClient
	bucket    string
	publicURL string
}

func NewR2Storage(cfg *config.R2Config) (*R2Storage, error) {
	r2Endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("falha ao configurar cliente R2: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(r2Endpoint)
		o.UsePathStyle = true
	})

	return &R2Storage{
		client:    client,
		presigner: s3.NewPresignClient(client),
		bucket:    cfg.Bucket,
		publicURL: cfg.PublicURL,
	}, nil
}

func (s *R2Storage) GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, ttl time.Duration) (string, error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}
	req, err := s.presigner.PresignPutObject(ctx, input, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("falha ao gerar presigned upload URL: %w", err)
	}
	return req.URL, nil
}

func (s *R2Storage) HeadObject(ctx context.Context, key string) (*ObjectMetadata, error) {
	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var respErr interface{ HTTPStatusCode() int }
		if errors.As(err, &respErr) && respErr.HTTPStatusCode() == http.StatusNotFound {
			return nil, ErrObjectNotFound
		}
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && (apiErr.ErrorCode() == "NotFound" || apiErr.ErrorCode() == "NoSuchKey") {
			return nil, ErrObjectNotFound
		}
		return nil, fmt.Errorf("falha ao consultar metadados no R2: %w", err)
	}

	return &ObjectMetadata{
		SizeBytes: aws.ToInt64(out.ContentLength),
		MimeType:  aws.ToString(out.ContentType),
	}, nil
}

func (s *R2Storage) GenerateSignedURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	req, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("falha ao gerar signed URL: %w", err)
	}
	return req.URL, nil
}

func (s *R2Storage) DeleteObject(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("falha ao deletar objeto R2: %w", err)
	}
	return nil
}
