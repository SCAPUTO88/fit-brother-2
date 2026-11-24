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
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	sqlxDB := sqlx.NewDb(mockDB, "postgres")
	repo := postgres.NewStudentRepository(sqlxDB)
	ctx := context.Background()

	t.Run("Deve executar o INSERT de Student com trainer_id e parametros isolados", func(t *testing.T) {
		trainerID := uuid.New()
		std := &student.Student{
			ID:           uuid.New(),
			UserID:       uuid.New(),
			TrainerID:    &trainerID,
			CPFEncrypted: "AES-RESULT-123",
			CPFHash:      "HASH_RESULT",
			CreatedAt:    time.Now(),
		}

		queryExpect := `INSERT INTO students \(id, user_id, trainer_id, cpf_encrypted, cpf_hash, created_at\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\)`

		mock.ExpectExec(queryExpect).
			WithArgs(std.ID, std.UserID, std.TrainerID, std.CPFEncrypted, std.CPFHash, std.CreatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, std)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Deve filtrar ListByTrainer pelo trainer_id autenticado", func(t *testing.T) {
		trainerID := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "full_name", "email"}).
			AddRow(uuid.New().String(), "Aluno Um", "aluno1@fit.com")

		mock.ExpectQuery(`SELECT s\.id, u\.full_name, u\.email FROM students s INNER JOIN users u ON s\.user_id = u\.id WHERE s\.trainer_id = \$1 AND u\.deleted_at IS NULL AND u\.is_active = true ORDER BY u\.full_name ASC`).
			WithArgs(trainerID).
			WillReturnRows(rows)

		items, err := repo.ListByTrainer(ctx, trainerID)
		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, "Aluno Um", items[0].FullName)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
