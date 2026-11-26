package storage

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainStorage "github.com/SCAPUTO88/fit-brother/internal/domain/storage"
)

func TestLocalStorageLifecycleAndSecurity(t *testing.T) {
	dir := t.TempDir()
	now := time.Unix(1700000000, 0)
	store, err := NewLocalStorage(dir, "http://localhost:8080", []byte("local-signing-secret-32-bytes!!"))
	require.NoError(t, err)
	store.nowFunc = func() time.Time { return now }

	ctx := context.Background()
	key := "exercises/ex-1/demo.mp4"

	t.Run("Rejeita path traversal e chaves invalidas", func(t *testing.T) {
		_, err := store.GeneratePresignedUploadURL(ctx, "../../etc/passwd", "video/mp4", 5*time.Minute)
		require.Error(t, err)

		_, err = store.HeadObject(ctx, "/abs/path.mp4")
		require.Error(t, err)

		_, err = store.GenerateSignedURL(ctx, "exercises/ex-1/demo.mp4.meta.json", 5*time.Minute)
		require.Error(t, err)
	})

	t.Run("Fluxo completo de upload assinado, HeadObject, playback e DeleteObject", func(t *testing.T) {
		uploadURL, err := store.GeneratePresignedUploadURL(ctx, key, "video/mp4", 5*time.Minute)
		require.NoError(t, err)

		parsedUpload, err := url.Parse(uploadURL)
		require.NoError(t, err)
		q := parsedUpload.Query()

		require.NoError(t, store.VerifySignedRequest("put", q.Get("key"), q.Get("content_type"), q.Get("expires"), q.Get("sig")))
		require.Error(t, store.VerifySignedRequest("put", q.Get("key"), "text/plain", q.Get("expires"), q.Get("sig")))
		require.Error(t, store.VerifySignedRequest("get", q.Get("key"), "", q.Get("expires"), q.Get("sig")))

		payload := []byte("fake-mp4-video-content")
		written, err := store.SaveObject(key, "video/mp4", bytes.NewReader(payload), 1024)
		require.NoError(t, err)
		assert.Equal(t, int64(len(payload)), written)

		meta, err := store.HeadObject(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, int64(len(payload)), meta.SizeBytes)
		assert.Equal(t, "video/mp4", meta.MimeType)

		signedURL, err := store.GenerateSignedURL(ctx, key, 5*time.Minute)
		require.NoError(t, err)
		parsedGet, err := url.Parse(signedURL)
		require.NoError(t, err)
		gq := parsedGet.Query()
		require.NoError(t, store.VerifySignedRequest("get", gq.Get("key"), "", gq.Get("expires"), gq.Get("sig")))

		// Expiracao da URL
		now = now.Add(6 * time.Minute)
		err = store.VerifySignedRequest("get", gq.Get("key"), "", gq.Get("expires"), gq.Get("sig"))
		assert.True(t, errors.Is(err, ErrInvalidSignature))

		require.NoError(t, store.DeleteObject(ctx, key))
		_, err = store.HeadObject(ctx, key)
		assert.True(t, errors.Is(err, domainStorage.ErrObjectNotFound))
	})

	t.Run("Rejeita arquivo acima do limite maximo", func(t *testing.T) {
		_, err := store.SaveObject("exercises/ex-1/huge.mp4", "video/mp4", bytes.NewReader([]byte("0123456789")), 5)
		assert.True(t, errors.Is(err, ErrFileTooLarge))
	})
}

func TestDisabledStorageReturnsErrStorageDisabled(t *testing.T) {
	ctx := context.Background()
	store := NewDisabledStorage()

	_, err := store.GeneratePresignedUploadURL(ctx, "a.mp4", "video/mp4", time.Minute)
	assert.True(t, errors.Is(err, domainStorage.ErrStorageDisabled))

	_, err = store.HeadObject(ctx, "a.mp4")
	assert.True(t, errors.Is(err, domainStorage.ErrStorageDisabled))

	_, err = store.GenerateSignedURL(ctx, "a.mp4", time.Minute)
	assert.True(t, errors.Is(err, domainStorage.ErrStorageDisabled))

	err = store.DeleteObject(ctx, "a.mp4")
	assert.True(t, errors.Is(err, domainStorage.ErrStorageDisabled))
}
