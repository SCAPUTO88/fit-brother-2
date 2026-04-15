//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/golang-migrate/migrate/v4"
	pgx5 "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/jmoiron/sqlx"
)

func main() {
	dir := "up"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://fitbrother:secret@localhost:5432/fitbrother_db?sslmode=disable"
	}

	db, err := sqlx.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("falha ao conectar ao banco: %v", err)
	}
	defer db.Close()

	driver, err := pgx5.WithInstance(db.DB, &pgx5.Config{})
	if err != nil {
		log.Fatalf("falha ao criar driver de migrate: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "pgx5", driver)
	if err != nil {
		log.Fatalf("falha ao inicializar migrate: %v", err)
	}

	switch dir {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate up falhou: %v", err)
		}
		fmt.Println("✅ Migrations aplicadas com sucesso!")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate down falhou: %v", err)
		}
		fmt.Println("✅ Rollback aplicado com sucesso!")
	default:
		log.Fatalf("argumento inválido: use 'up' ou 'down'")
	}
}
