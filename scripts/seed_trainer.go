//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/joho/godotenv"

	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/config"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
)

func main() {
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := sqlx.Open("pgx", cfg.Database.URL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()

	passwordSvc, err := crypto.NewPasswordService(cfg.Crypto.PasswordPepper)
	if err != nil {
		log.Fatalf("crypto: %v", err)
	}

	email := "trainer@fitbrother.com"
	password := "Senha123!"

	if len(os.Args) > 1 {
		email = os.Args[1]
	}
	if len(os.Args) > 2 {
		password = os.Args[2]
	}

	hash, err := passwordSvc.Hash(password)
	if err != nil {
		log.Fatalf("hash: %v", err)
	}

	id := uuid.New()
	_, err = db.Exec(`
		INSERT INTO users (id, email, full_name, password_hash, role, is_active)
		VALUES ($1, $2, $3, $4, 'TRAINER', true)
		ON CONFLICT DO NOTHING
	`, id, email, "Trainer Principal", hash)
	if err != nil {
		log.Fatalf("insert: %v", err)
	}

	// Atualiza o hash se o usuário já existia
	_, err = db.Exec(`UPDATE users SET password_hash = $1, role = 'TRAINER' WHERE email = $2`, hash, email)
	if err != nil {
		log.Fatalf("update: %v", err)
	}

	fmt.Printf("✅ Trainer criado/atualizado: %s | senha: %s\n", email, password)
}
