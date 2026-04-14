package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/adapters/repository/postgres"
	"github.com/SCAPUTO88/fit-brother/internal/domain/student"
)

func TestStudentRepository_Create(t *testing.T) {
	// Setup do Mock Database
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	sqlxDB := sqlx.NewDb(mockDB, "postgres")
	repo := postgres.NewStudentRepository(sqlxDB)

	ctx := context.TODO()

	t.Run("Deve executar o INSERT de Student com os parametros isolados (Prepared statement)", func(t *testing.T) {
		std := &student.Student{
			ID:           uuid.New(),
			UserID:       uuid.New(),
			CPFEncrypted: "AES-RESULT-123",
			CPFHash:      "HASH_RESULT",
			CreatedAt:    time.Now(),
		}

		// A Regex abaixo valida que o sqlx/driver usarão Placeholders ao invés de String Interpolation direta, barrando SQL Injection
		queryExpect := `INSERT INTO students \(id, user_id, cpf_encrypted, cpf_hash, created_at\) VALUES \(\$1, \$2, \$3, \$4, \$5\)`

		mock.ExpectExec(queryExpect).
			WithArgs(std.ID, std.UserID, std.CPFEncrypted, std.CPFHash, std.CreatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, std)
		assert.NoError(t, err)

		// Verifica se o código de fato ativou/bateu nas exatas exigencias do nosso mock preventivo de SQL Injection
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}
