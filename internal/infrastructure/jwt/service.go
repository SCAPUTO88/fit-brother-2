package jwt

import (
	"crypto/rsa"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims define o payload do nosso JWT.
type Claims struct {
	UserID    uuid.UUID `json:"sub"`
	Role      string    `json:"role"`
	SessionID string    `json:"sid"`
	jwt.RegisteredClaims
}

// Service gerencia a criação e validação de tokens JWT usando algoritmo assimétrico RS256.
type Service struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
}

// NewService carrega as chaves RSA dos caminhos fornecidos.
func NewService(privateKeyPath, publicKeyPath, issuer string) (*Service, error) {
	privBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, errors.New("falha ao ler private key: " + err.Error())
	}
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privBytes)
	if err != nil {
		return nil, errors.New("falha ao fazer parse da private key: " + err.Error())
	}

	pubBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, errors.New("falha ao ler public key: " + err.Error())
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

// GenerateAccessToken assina um novo JWT válido por `ttl`.
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
			ID:        uuid.NewString(), // jti único
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

// Verify decodifica um token e valida sua assinatura usando a Public Key.
func (s *Service) Verify(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validar se o método de assinatura é o esperado
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
