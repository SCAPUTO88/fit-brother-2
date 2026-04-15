//go:build ignore

package main

import (
	"fmt"
	"log"

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

	fmt.Printf("Pepper (hex): %x\n", []byte(cfg.Crypto.PasswordPepper))
	fmt.Printf("Pepper (str): %q\n", cfg.Crypto.PasswordPepper)

	svc, _ := crypto.NewPasswordService(cfg.Crypto.PasswordPepper)

	hash, _ := svc.Hash("Senha123!")
	fmt.Printf("Hash gerado: %s\n", hash)
	fmt.Printf("Verificação: %v\n", svc.Check("Senha123!", hash))
}
