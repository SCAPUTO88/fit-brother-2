package storage

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	domainStorage "github.com/SCAPUTO88/fit-brother/internal/domain/storage"
)

var (
	_ domainStorage.Port = (*LocalStorage)(nil)

	ErrInvalidSignature = errors.New("assinatura ou validade da URL de storage inválida")
	ErrInvalidKey       = errors.New("chave de objeto inválida")
	ErrFileTooLarge     = errors.New("arquivo excede o limite máximo permitido")
)

type localMeta struct {
	SizeBytes int64  `json:"size_bytes"`
	MimeType  string `json:"mime_type"`
}

type LocalStorage struct {
	baseDir       string
	baseURL       string
	signingSecret []byte
	nowFunc       func() time.Time
}

func NewLocalStorage(baseDir, baseURL string, signingSecret []byte) (*LocalStorage, error) {
	if strings.TrimSpace(baseDir) == "" {
		return nil, errors.New("diretório base do local storage é obrigatório")
	}
	if len(signingSecret) == 0 {
		return nil, errors.New("segredo de assinatura do local storage é obrigatório")
	}

	absDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("falha ao resolver diretório do local storage: %w", err)
	}
	if err := os.MkdirAll(absDir, 0o750); err != nil {
		return nil, fmt.Errorf("falha ao criar diretório do local storage: %w", err)
	}

	secretCopy := make([]byte, len(signingSecret))
	copy(secretCopy, signingSecret)

	return &LocalStorage{
		baseDir:       absDir,
		baseURL:       strings.TrimRight(baseURL, "/"),
		signingSecret: secretCopy,
		nowFunc:       time.Now,
	}, nil
}

func (s *LocalStorage) GeneratePresignedUploadURL(_ context.Context, key string, contentType string, ttl time.Duration) (string, error) {
	if _, err := s.resolveSafePath(key); err != nil {
		return "", err
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	expires := s.nowFunc().Add(ttl).Unix()
	sig := s.sign("put", key, contentType, expires)

	q := url.Values{}
	q.Set("key", key)
	q.Set("op", "put")
	if contentType != "" {
		q.Set("content_type", contentType)
	}
	q.Set("expires", strconv.FormatInt(expires, 10))
	q.Set("sig", sig)

	return s.baseURL + "/v1/storage/local?" + q.Encode(), nil
}

func (s *LocalStorage) GenerateSignedURL(_ context.Context, key string, ttl time.Duration) (string, error) {
	if _, err := s.resolveSafePath(key); err != nil {
		return "", err
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	expires := s.nowFunc().Add(ttl).Unix()
	sig := s.sign("get", key, "", expires)

	q := url.Values{}
	q.Set("key", key)
	q.Set("op", "get")
	q.Set("expires", strconv.FormatInt(expires, 10))
	q.Set("sig", sig)

	return s.baseURL + "/v1/storage/local?" + q.Encode(), nil
}

func (s *LocalStorage) HeadObject(_ context.Context, key string) (*domainStorage.ObjectMetadata, error) {
	targetPath, err := s.resolveSafePath(key)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(targetPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, domainStorage.ErrObjectNotFound
		}
		return nil, fmt.Errorf("falha ao consultar objeto local: %w", err)
	}
	if info.IsDir() {
		return nil, domainStorage.ErrObjectNotFound
	}

	meta := &domainStorage.ObjectMetadata{
		SizeBytes: info.Size(),
	}
	if storedMeta, err := s.readMeta(targetPath); err == nil && storedMeta != nil {
		meta.MimeType = storedMeta.MimeType
	}

	return meta, nil
}

func (s *LocalStorage) DeleteObject(_ context.Context, key string) error {
	targetPath, err := s.resolveSafePath(key)
	if err != nil {
		return err
	}

	if err := os.Remove(targetPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("falha ao remover objeto local: %w", err)
	}
	if err := os.Remove(metaPath(targetPath)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("falha ao remover metadados locais: %w", err)
	}

	return nil
}

func (s *LocalStorage) VerifySignedRequest(op, key, contentType, expiresStr, sigHex string) error {
	if op != "put" && op != "get" {
		return ErrInvalidSignature
	}
	if _, err := s.resolveSafePath(key); err != nil {
		return ErrInvalidKey
	}
	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil || expires < s.nowFunc().Unix() {
		return ErrInvalidSignature
	}

	expectedSig := s.sign(op, key, contentType, expires)
	expectedBytes, err := hex.DecodeString(expectedSig)
	if err != nil {
		return ErrInvalidSignature
	}
	providedBytes, err := hex.DecodeString(sigHex)
	if err != nil {
		return ErrInvalidSignature
	}

	if !hmac.Equal(expectedBytes, providedBytes) {
		return ErrInvalidSignature
	}
	return nil
}

func (s *LocalStorage) SaveObject(key, contentType string, body io.Reader, maxBytes int64) (int64, error) {
	targetPath, err := s.resolveSafePath(key)
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o750); err != nil {
		return 0, fmt.Errorf("falha ao criar diretório do objeto: %w", err)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(targetPath), ".upload-*")
	if err != nil {
		return 0, fmt.Errorf("falha ao criar arquivo temporário: %w", err)
	}
	tmpPath := tmpFile.Name()
	cleanupTmp := true
	defer func() {
		if cleanupTmp {
			_ = tmpFile.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	reader := body
	if maxBytes > 0 {
		reader = io.LimitReader(body, maxBytes+1)
	}

	written, err := io.Copy(tmpFile, reader)
	if err != nil {
		return 0, fmt.Errorf("falha ao gravar arquivo local: %w", err)
	}
	if maxBytes > 0 && written > maxBytes {
		return 0, ErrFileTooLarge
	}
	if err := tmpFile.Close(); err != nil {
		return 0, fmt.Errorf("falha ao finalizar arquivo local: %w", err)
	}

	if err := os.Rename(tmpPath, targetPath); err != nil {
		return 0, fmt.Errorf("falha ao mover arquivo para destino final: %w", err)
	}
	cleanupTmp = false

	metaBytes, err := json.Marshal(localMeta{
		SizeBytes: written,
		MimeType:  contentType,
	})
	if err != nil {
		return 0, fmt.Errorf("falha ao serializar metadados locais: %w", err)
	}
	if err := os.WriteFile(metaPath(targetPath), metaBytes, 0o640); err != nil {
		return 0, fmt.Errorf("falha ao gravar metadados locais: %w", err)
	}

	return written, nil
}

func (s *LocalStorage) OpenObject(key string) (*os.File, *domainStorage.ObjectMetadata, os.FileInfo, error) {
	targetPath, err := s.resolveSafePath(key)
	if err != nil {
		return nil, nil, nil, err
	}

	info, err := os.Stat(targetPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, nil, domainStorage.ErrObjectNotFound
		}
		return nil, nil, nil, err
	}
	if info.IsDir() {
		return nil, nil, nil, domainStorage.ErrObjectNotFound
	}

	f, err := os.Open(targetPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, nil, domainStorage.ErrObjectNotFound
		}
		return nil, nil, nil, err
	}

	meta := &domainStorage.ObjectMetadata{SizeBytes: info.Size()}
	if stored, err := s.readMeta(targetPath); err == nil && stored != nil {
		meta.MimeType = stored.MimeType
	}

	return f, meta, info, nil
}

func (s *LocalStorage) readMeta(targetPath string) (*localMeta, error) {
	data, err := os.ReadFile(metaPath(targetPath))
	if err != nil {
		return nil, err
	}
	var m localMeta
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *LocalStorage) sign(op, key, contentType string, expires int64) string {
	mac := hmac.New(sha256.New, s.signingSecret)
	payload := strings.ToUpper(op) + "\n" + key + "\n" + contentType + "\n" + strconv.FormatInt(expires, 10)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *LocalStorage) resolveSafePath(key string) (string, error) {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" || strings.ContainsRune(trimmed, '\\') || strings.HasPrefix(trimmed, "/") {
		return "", ErrInvalidKey
	}

	cleaned := path.Clean(trimmed)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") {
		return "", ErrInvalidKey
	}

	relOS := filepath.FromSlash(cleaned)
	fullPath := filepath.Clean(filepath.Join(s.baseDir, relOS))
	prefix := s.baseDir + string(filepath.Separator)
	if !strings.HasPrefix(fullPath, prefix) {
		return "", ErrInvalidKey
	}
	if strings.HasSuffix(fullPath, ".meta.json") {
		return "", ErrInvalidKey
	}

	return fullPath, nil
}

func metaPath(targetPath string) string {
	return targetPath + ".meta.json"
}
