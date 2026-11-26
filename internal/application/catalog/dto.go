package catalog

import (
	"time"

	domainCatalog "github.com/SCAPUTO88/fit-brother/internal/domain/catalog"
)

type FiltersOutput struct {
	MuscleGroups []*domainCatalog.MuscleGroup   `json:"muscle_groups"`
	Equipments   []*domainCatalog.Equipment     `json:"equipments"`
	Categories   []domainCatalog.CategoryOption `json:"categories"`
}

type MediaDTO struct {
	ID        string    `json:"id"`
	MediaType string    `json:"media_type"`
	MimeType  string    `json:"mime_type"`
	Filename  string    `json:"filename"`
	SizeBytes int64     `json:"size_bytes"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type ExerciseDTO struct {
	ID               string   `json:"id"`
	OrganizationID   *string  `json:"organization_id,omitempty"`
	ExternalSource   string   `json:"external_source"`
	ExternalID       *string  `json:"external_id,omitempty"`
	NamePT           string   `json:"name_pt"`
	NameEN           string   `json:"name_en,omitempty"`
	PrimaryMuscle    string   `json:"primary_muscle"`
	PrimaryMusclePT  string   `json:"primary_muscle_pt"`
	SecondaryMuscles []string `json:"secondary_muscles"`
	Equipment        string   `json:"equipment"`
	EquipmentPT      string   `json:"equipment_pt"`
	Category         string   `json:"category"`
	CategoryPT       string   `json:"category_pt"`
	Level            string   `json:"level"`
	Force            string   `json:"force,omitempty"`
	Mechanic         string   `json:"mechanic,omitempty"`
	LicenseInfo      string   `json:"license_info"`
	IsCustom         bool     `json:"is_custom"`
}

type ExerciseDetailOutput struct {
	ExerciseDTO
	InstructionsPT []string   `json:"instructions_pt"`
	InstructionsEN []string   `json:"instructions_en,omitempty"`
	Media          []MediaDTO `json:"media"`
}

type SearchResultOutput struct {
	Items  []ExerciseDTO `json:"items"`
	Total  int           `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

type CreateExerciseInput struct {
	OrganizationID   string   `json:"organization_id"`
	NamePT           string   `json:"name_pt"`
	NameEN           string   `json:"name_en"`
	PrimaryMuscle    string   `json:"primary_muscle"`
	SecondaryMuscles []string `json:"secondary_muscles"`
	Equipment        string   `json:"equipment"`
	Category         string   `json:"category"`
	Level            string   `json:"level"`
	Force            string   `json:"force"`
	Mechanic         string   `json:"mechanic"`
	InstructionsPT   []string `json:"instructions_pt"`
	ClientIP         string   `json:"-"`
}

type UpdateExerciseInput struct {
	NamePT           string   `json:"name_pt"`
	PrimaryMuscle    string   `json:"primary_muscle"`
	SecondaryMuscles []string `json:"secondary_muscles"`
	Equipment        string   `json:"equipment"`
	Category         string   `json:"category"`
	Level            string   `json:"level"`
	Force            string   `json:"force"`
	Mechanic         string   `json:"mechanic"`
	InstructionsPT   []string `json:"instructions_pt"`
	ClientIP         string   `json:"-"`
}

type RequestMediaUploadInput struct {
	Filename  string `json:"filename"`
	MimeType  string `json:"mime_type"`
	MediaType string `json:"media_type"`
}

type RequestMediaUploadOutput struct {
	MediaID      string `json:"media_id"`
	PresignedURL string `json:"presigned_url"`
	StorageKey   string `json:"storage_key"`
	MediaType    string `json:"media_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type ConfirmMediaUploadInput struct {
	StorageKey string `json:"storage_key"`
	Filename   string `json:"filename"`
	MimeType   string `json:"mime_type"`
	SizeBytes  int64  `json:"size_bytes"`
	ClientIP   string `json:"-"`
}

type SignedMediaURLOutput struct {
	SignedURL string `json:"signed_url"`
	MediaType string `json:"media_type"`
	ExpiresIn int    `json:"expires_in"`
}

type ImportResultOutput struct {
	Processed int `json:"processed"`
	Inserted  int `json:"inserted"`
	Updated   int `json:"updated"`
	Skipped   int `json:"skipped"`
}
