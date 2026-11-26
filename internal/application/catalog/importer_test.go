package catalog_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/application/catalog"
	domainCatalog "github.com/SCAPUTO88/fit-brother/internal/domain/catalog"
	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
)

func TestImporterIdempotentAndPTBRTranslation(t *testing.T) {
	ctx := context.Background()
	repo := newMemCatalogRepo()
	imp := catalog.NewImporter(repo)

	sampleJSON := `[
		{
			"id": "Barbell_Bench_Press_-_Medium_Grip",
			"name": "Barbell Bench Press - Medium Grip",
			"force": "push",
			"level": "beginner",
			"mechanic": "compound",
			"equipment": "barbell",
			"primaryMuscles": ["chest"],
			"secondaryMuscles": ["shoulders", "triceps"],
			"instructions": [
				"Lie back on a flat bench and bring the barbell slowly down to your middle chest."
			],
			"category": "strength"
		},
		{
			"id": "Incline_Dumbbell_Press",
			"name": "Incline Dumbbell Press",
			"force": "push",
			"level": "intermediate",
			"mechanic": "compound",
			"equipment": "dumbbell",
			"primaryMuscles": ["chest"],
			"secondaryMuscles": ["shoulders"],
			"instructions": [
				"Press dumbbells up from incline bench with control."
			],
			"category": "strength"
		},
		{
			"id": "",
			"name": "Invalid Record"
		}
	]`

	first, err := imp.ImportFreeExerciseDB(ctx, strings.NewReader(sampleJSON))
	require.NoError(t, err)
	assert.Equal(t, 2, first.Processed)
	assert.Equal(t, 2, first.Inserted)
	assert.Equal(t, 0, first.Updated)
	assert.Equal(t, 1, first.Skipped)

	// Segunda execução com os mesmos dados deve atualizar em vez de duplicar
	second, err := imp.ImportFreeExerciseDB(ctx, strings.NewReader(sampleJSON))
	require.NoError(t, err)
	assert.Equal(t, 2, second.Processed)
	assert.Equal(t, 0, second.Inserted)
	assert.Equal(t, 2, second.Updated)
	assert.Equal(t, 1, second.Skipped)
	assert.Len(t, repo.exercises, 2)

	var bench *domainCatalog.Exercise
	for _, ex := range repo.exercises {
		if ex.ExternalID != nil && *ex.ExternalID == "Barbell_Bench_Press_-_Medium_Grip" {
			bench = ex
		}
	}
	require.NotNil(t, bench)
	assert.Equal(t, "Supino Reto com Barra", bench.NamePT)
	assert.Equal(t, domainCatalog.LicenseUnlicense, bench.LicenseInfo)
	assert.NotEmpty(t, bench.InstructionsPT)
	assert.Contains(t, bench.InstructionsPT[0], "banco reto")
	assert.Contains(t, bench.InstructionsPT[0], "barra")

	var incline *domainCatalog.Exercise
	for _, ex := range repo.exercises {
		if ex.ExternalID != nil && *ex.ExternalID == "Incline_Dumbbell_Press" {
			incline = ex
		}
	}
	require.NotNil(t, incline)
	assert.Equal(t, "Supino Inclinado com Halteres", incline.NamePT)
	assert.Equal(t, "Peitoral", catalog.TranslateMusclePT(incline.PrimaryMuscle))
	assert.Equal(t, "Halteres", catalog.TranslateEquipmentPT(incline.Equipment))
}

func TestImporterRejectsMalformedJSON(t *testing.T) {
	ctx := context.Background()
	repo := newMemCatalogRepo()
	imp := catalog.NewImporter(repo)

	_, err := imp.ImportFreeExerciseDB(ctx, strings.NewReader("{invalid"))
	require.ErrorIs(t, err, errs.ErrInvalidInput)

	_, err = imp.ImportFreeExerciseDB(ctx, nil)
	require.ErrorIs(t, err, errs.ErrInvalidInput)

	assert.Equal(t, uuid.Nil, uuid.UUID{})
}
