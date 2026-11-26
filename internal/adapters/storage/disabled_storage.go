package storage

import (
	"context"
	"time"

	domainStorage "github.com/SCAPUTO88/fit-brother/internal/domain/storage"
)

var _ domainStorage.Port = (*DisabledStorage)(nil)

type DisabledStorage struct{}

func NewDisabledStorage() *DisabledStorage {
	return &DisabledStorage{}
}

func (s *DisabledStorage) GeneratePresignedUploadURL(context.Context, string, string, time.Duration) (string, error) {
	return "", domainStorage.ErrStorageDisabled
}

func (s *DisabledStorage) HeadObject(context.Context, string) (*domainStorage.ObjectMetadata, error) {
	return nil, domainStorage.ErrStorageDisabled
}

func (s *DisabledStorage) GenerateSignedURL(context.Context, string, time.Duration) (string, error) {
	return "", domainStorage.ErrStorageDisabled
}

func (s *DisabledStorage) DeleteObject(context.Context, string) error {
	return domainStorage.ErrStorageDisabled
}
