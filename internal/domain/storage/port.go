package storage

import (
	"context"
	"errors"
	"time"
)

var (
	ErrObjectNotFound  = errors.New("objeto não encontrado no storage")
	ErrStorageDisabled = errors.New("armazenamento de mídia desabilitado neste ambiente")
)

type ObjectMetadata struct {
	SizeBytes int64
	MimeType  string
}

type Port interface {
	GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, ttl time.Duration) (string, error)
	HeadObject(ctx context.Context, key string) (*ObjectMetadata, error)
	GenerateSignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
	DeleteObject(ctx context.Context, key string) error
}
