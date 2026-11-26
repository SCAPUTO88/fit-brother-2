package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/SCAPUTO88/fit-brother/internal/adapters/repository/postgres"
	appCatalog "github.com/SCAPUTO88/fit-brother/internal/application/catalog"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/config"
)

func main() {
	filePath := flag.String("file", "", "Caminho para o arquivo JSON do free-exercise-db")
	flag.Parse()

	if err := run(*filePath); err != nil {
		fmt.Fprintf(os.Stderr, "erro na importação do catálogo: %v\n", err)
		os.Exit(1)
	}
}

func run(rawPath string) error {
	cleaned := filepath.Clean(strings.TrimSpace(rawPath))
	if cleaned == "" || cleaned == "." {
		return fmt.Errorf("parâmetro -file é obrigatório")
	}
	if strings.ToLower(filepath.Ext(cleaned)) != ".json" {
		return fmt.Errorf("apenas arquivos .json são suportados")
	}

	absPath, err := filepath.Abs(cleaned)
	if err != nil {
		return fmt.Errorf("caminho de arquivo inválido: %w", err)
	}

	dir := filepath.Dir(absPath)
	base := filepath.Base(absPath)
	f, err := os.OpenInRoot(dir, base)
	if err != nil {
		return fmt.Errorf("falha ao abrir arquivo %s: %w", base, err)
	}
	defer func() {
		_ = f.Close()
	}()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("falha ao carregar configuração: %w", err)
	}

	db, err := sqlx.Connect("pgx", cfg.Database.URL)
	if err != nil {
		return fmt.Errorf("falha ao conectar no Postgres: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	repo := postgres.NewCatalogRepository(db)
	importer := appCatalog.NewImporter(repo)
	res, err := importer.ImportFreeExerciseDB(ctx, f)
	if err != nil {
		return err
	}

	fmt.Printf(
		"Importação concluída: processed=%d inserted=%d updated=%d skipped=%d\n",
		res.Processed,
		res.Inserted,
		res.Updated,
		res.Skipped,
	)
	return nil
}
