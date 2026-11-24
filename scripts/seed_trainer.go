//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	passwordPepper := os.Getenv("PASSWORD_PEPPER")
	if passwordPepper == "" {
		log.Fatal("PASSWORD_PEPPER is required")
	}

	email := os.Getenv("SEED_TRAINER_EMAIL")
	if email == "" {
		email = "trainer@fitbrother.com"
	}
	if len(os.Args) > 1 {
		email = os.Args[1]
	}

	password := os.Getenv("SEED_TRAINER_PASSWORD")
	if len(os.Args) > 2 {
		password = os.Args[2]
	}
	if password == "" {
		log.Fatal("SEED_TRAINER_PASSWORD env var or second CLI argument is required")
	}

	db, err := sqlx.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()

	passwordSvc, err := crypto.NewPasswordService(passwordPepper)
	if err != nil {
		log.Fatalf("crypto: %v", err)
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

	_, err = db.Exec(`UPDATE users SET password_hash = $1, role = 'TRAINER', is_active = true, deleted_at = NULL WHERE email = $2`, hash, email)
	if err != nil {
		log.Fatalf("update: %v", err)
	}

	fmt.Printf("✅ Trainer criado/atualizado: %s\n", email)
}
