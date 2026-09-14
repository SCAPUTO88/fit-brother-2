package jwt

import (
	"crypto/rsa"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID    uuid.UUID `json:"sub"`
	Role      string    `json:"role"`
	SessionID string    `json:"sid"`
	jwt.RegisteredClaims
}

type Service struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
}

func NewService(privateKeyPath, publicKeyPath, issuer string) (*Service, error) {
	privBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, errors.New("falha ao ler private key: " + err.Error())
	}

	pubBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, errors.New("falha ao ler public key: " + err.Error())
	}

	return newService(privBytes, pubBytes, issuer)
}

func NewServiceFromPEM(privateKeyPEM, publicKeyPEM, issuer string) (*Service, error) {
	return newService([]byte(privateKeyPEM), []byte(publicKeyPEM), issuer)
}

func newService(privBytes, pubBytes []byte, issuer string) (*Service, error) {
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privBytes)
	if err != nil {
		return nil, errors.New("falha ao fazer parse da private key: " + err.Error())
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		return nil, errors.New("falha ao fazer parse da public key: " + err.Error())
	}

	return &Service{
		privateKey: privKey,
		publicKey:  pubKey,
		issuer:     issuer,
	}, nil
}

func (s *Service) GenerateAccessToken(userID uuid.UUID, role string, sessionID string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Role:      role,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        uuid.NewString(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

func (s *Service) Verify(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("método de assinatura inesperado")
		}
		return s.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("token inválido")
	}

	return claims, nil
}
