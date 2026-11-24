//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/config"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("sem .env:", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	testPassword := os.Getenv("DEBUG_TEST_PASSWORD")
	if len(os.Args) > 1 {
		testPassword = os.Args[1]
	}
	if testPassword == "" {
		log.Fatal("informe DEBUG_TEST_PASSWORD ou passe a senha de teste via argumento")
	}

	svc, err := crypto.NewPasswordService(cfg.Crypto.PasswordPepper)
	if err != nil {
		log.Fatalf("crypto: %v", err)
	}

	hash, err := svc.Hash(testPassword)
	if err != nil {
		log.Fatalf("hash: %v", err)
	}
	fmt.Printf("Hash gerado com sucesso (comprimento=%d)\n", len(hash))
	fmt.Printf("Verificação: %v\n", svc.Check(testPassword, hash))
}
