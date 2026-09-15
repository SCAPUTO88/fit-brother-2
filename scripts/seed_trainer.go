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

	db, err := sqlx.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()

	passwordSvc, err := crypto.NewPasswordService(passwordPepper)
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

	fmt.Printf("✅ Trainer criado/atualizado: %s\n", email)
}
