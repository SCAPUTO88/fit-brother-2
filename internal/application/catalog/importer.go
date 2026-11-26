package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"

	domainCatalog "github.com/SCAPUTO88/fit-brother/internal/domain/catalog"
	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
)

// FreeExerciseDBRecord representa o schema público do projeto yuhonas/free-exercise-db (The Unlicense).
type FreeExerciseDBRecord struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Force            string   `json:"force"`
	Level            string   `json:"level"`
	Mechanic         string   `json:"mechanic"`
	Equipment        string   `json:"equipment"`
	PrimaryMuscles   []string `json:"primaryMuscles"`
	SecondaryMuscles []string `json:"secondaryMuscles"`
	Instructions     []string `json:"instructions"`
	Category         string   `json:"category"`
	Images           []string `json:"images"`
}

type Importer struct {
	repo domainCatalog.Repository
}

func NewImporter(repo domainCatalog.Repository) *Importer {
	return &Importer{repo: repo}
}

func (imp *Importer) ImportFreeExerciseDB(ctx context.Context, r io.Reader) (*ImportResultOutput, error) {
	if r == nil {
		return nil, fmt.Errorf("%w: fonte JSON de exercícios é obrigatória", errs.ErrInvalidInput)
	}

	var records []FreeExerciseDBRecord
	if err := json.NewDecoder(r).Decode(&records); err != nil {
		return nil, fmt.Errorf("%w: JSON do catálogo free-exercise-db inválido", errs.ErrInvalidInput)
	}

	result := &ImportResultOutput{}
	for _, rec := range records {
		extID := strings.TrimSpace(rec.ID)
		nameEN := strings.TrimSpace(rec.Name)
		if extID == "" || nameEN == "" || len(rec.PrimaryMuscles) == 0 {
			result.Skipped++
			continue
		}

		primaryMuscle := domainCatalog.NormalizeCode(rec.PrimaryMuscles[0])
		if primaryMuscle == "" {
			result.Skipped++
			continue
		}

		secMuscles := make([]string, 0, len(rec.SecondaryMuscles))
		for _, m := range rec.SecondaryMuscles {
			if norm := domainCatalog.NormalizeCode(m); norm != "" {
				secMuscles = append(secMuscles, norm)
			}
		}

		equipment := domainCatalog.NormalizeCode(rec.Equipment)
		if equipment == "" {
			equipment = "body only"
		}
		category := domainCatalog.NormalizeCode(rec.Category)
		if category == "" {
			category = "strength"
		}
		level := domainCatalog.NormalizeCode(rec.Level)
		if level == "" {
			level = "intermediate"
		}

		ex := &domainCatalog.Exercise{
			ID:               uuid.New(),
			OrganizationID:   nil,
			CreatedByUserID:  nil,
			ExternalSource:   domainCatalog.SourceFreeExerciseDB,
			ExternalID:       &extID,
			NamePT:           TranslateExerciseNamePT(nameEN, primaryMuscle, equipment),
			NameEN:           nameEN,
			PrimaryMuscle:    primaryMuscle,
			SecondaryMuscles: secMuscles,
			Equipment:        equipment,
			Category:         category,
			Level:            level,
			Force:            domainCatalog.NormalizeCode(rec.Force),
			Mechanic:         domainCatalog.NormalizeCode(rec.Mechanic),
			InstructionsPT:   TranslateInstructionsPT(rec.Instructions),
			InstructionsEN:   cleanInstructions(rec.Instructions),
			LicenseInfo:      domainCatalog.LicenseUnlicense,
			IsActive:         true,
		}

		created, err := imp.repo.UpsertGlobalExercise(ctx, ex)
		if err != nil {
			return nil, fmt.Errorf("falha ao importar exercício %s: %w", extID, err)
		}
		result.Processed++
		if created {
			result.Inserted++
		} else {
			result.Updated++
		}
	}

	return result, nil
}

func cleanInstructions(items []string) []string {
	out := make([]string, 0, len(items))
	for _, s := range items {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
