package workout

import "time"

type TemplateExerciseInput struct {
	CatalogExerciseID string `json:"catalog_exercise_id"`
	MediaID           string `json:"media_id"`
	Name              string `json:"name"`
	PrimaryMuscle     string `json:"primary_muscle"`
	Equipment         string `json:"equipment"`
	Sets              int    `json:"sets" binding:"min=0"`
	Reps              string `json:"reps"`
	RestSeconds       int    `json:"rest_seconds" binding:"min=0"`
	Notes             string `json:"notes"`
	Position          int    `json:"position"`
}

type TemplateSessionInput struct {
	Title     string                  `json:"title" binding:"required,min=2"`
	DayOfWeek int                     `json:"day_of_week" binding:"min=0,max=6"`
	Notes     string                  `json:"notes"`
	Position  int                     `json:"position"`
	Exercises []TemplateExerciseInput `json:"exercises"`
}

type CreateTemplateInput struct {
	OrganizationID string                 `json:"organization_id"`
	Title          string                 `json:"title" binding:"required,min=3"`
	Description    string                 `json:"description"`
	Goal           string                 `json:"goal"`
	Sessions       []TemplateSessionInput `json:"sessions"`
	ClientIP       string                 `json:"-"`
}

type UpdateTemplateInput struct {
	Title       string                 `json:"title" binding:"required,min=3"`
	Description string                 `json:"description"`
	Goal        string                 `json:"goal"`
	IsArchived  *bool                  `json:"is_archived"`
	Sessions    []TemplateSessionInput `json:"sessions"`
	ClientIP    string                 `json:"-"`
}

type CloneTemplateInput struct {
	Title    string `json:"title"`
	ClientIP string `json:"-"`
}

type AssignTemplateInput struct {
	StudentID        string `json:"student_id" binding:"required,uuid"`
	WorkoutVersionID string `json:"workout_version_id"`
	ValidFrom        string `json:"valid_from"`
	ValidUntil       string `json:"valid_until"`
	Notes            string `json:"notes"`
	PublishIfNeeded  *bool  `json:"publish_if_needed"`
	ClientIP         string `json:"-"`
}

type UpdateAssignmentStatusInput struct {
	Status   string `json:"status" binding:"required"`
	ClientIP string `json:"-"`
}

type TemplateExerciseOutput struct {
	ID                string  `json:"id"`
	CatalogExerciseID *string `json:"catalog_exercise_id,omitempty"`
	LegacyExerciseID  *string `json:"legacy_exercise_id,omitempty"`
	MediaID           *string `json:"media_id,omitempty"`
	Name              string  `json:"name"`
	PrimaryMuscle     string  `json:"primary_muscle,omitempty"`
	Equipment         string  `json:"equipment,omitempty"`
	Sets              int     `json:"sets"`
	Reps              string  `json:"reps"`
	RestSeconds       int     `json:"rest_seconds"`
	Notes             string  `json:"notes,omitempty"`
	Position          int     `json:"position"`
	HasMedia          bool    `json:"has_media"`
}

type TemplateSessionOutput struct {
	ID        string                   `json:"id"`
	Title     string                   `json:"title"`
	DayOfWeek int                      `json:"day_of_week"`
	Notes     string                   `json:"notes,omitempty"`
	Position  int                      `json:"position"`
	Exercises []TemplateExerciseOutput `json:"exercises"`
}

type WorkoutStructureOutput struct {
	Sessions []TemplateSessionOutput `json:"sessions"`
}

type TemplateSummaryOutput struct {
	ID                   string    `json:"id"`
	OrganizationID       string    `json:"organization_id"`
	TrainerUserID        string    `json:"trainer_user_id"`
	Title                string    `json:"title"`
	Description          string    `json:"description"`
	Goal                 string    `json:"goal"`
	ClonedFromTemplateID *string   `json:"cloned_from_template_id,omitempty"`
	IsArchived           bool      `json:"is_archived"`
	LatestVersionNumber  int       `json:"latest_version_number"`
	LatestVersionID      *string   `json:"latest_version_id,omitempty"`
	SessionCount         int       `json:"session_count"`
	ExerciseCount        int       `json:"exercise_count"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type VersionSummaryOutput struct {
	ID            string    `json:"id"`
	TemplateID    string    `json:"template_id"`
	VersionNumber int       `json:"version_number"`
	Title         string    `json:"title"`
	Goal          string    `json:"goal"`
	SessionCount  int       `json:"session_count"`
	PublishedAt   time.Time `json:"published_at"`
}

type TemplateDetailOutput struct {
	TemplateSummaryOutput
	DraftStructure WorkoutStructureOutput `json:"draft_structure"`
	Versions       []VersionSummaryOutput `json:"versions"`
}

type VersionDetailOutput struct {
	ID                string                 `json:"id"`
	TemplateID        string                 `json:"template_id"`
	VersionNumber     int                    `json:"version_number"`
	Title             string                 `json:"title"`
	Description       string                 `json:"description"`
	Goal              string                 `json:"goal"`
	SnapshotStructure WorkoutStructureOutput `json:"snapshot_structure"`
	PublishedByUserID *string                `json:"published_by_user_id,omitempty"`
	PublishedAt       time.Time              `json:"published_at"`
}

type AssignmentSummaryOutput struct {
	ID                  string    `json:"id"`
	OrganizationID      string    `json:"organization_id"`
	TemplateID          string    `json:"template_id"`
	WorkoutVersionID    string    `json:"workout_version_id"`
	StudentID           string    `json:"student_id"`
	StudentName         string    `json:"student_name"`
	TrainerUserID       string    `json:"trainer_user_id"`
	LegacyWorkoutPlanID *string   `json:"legacy_workout_plan_id,omitempty"`
	TemplateTitle       string    `json:"template_title"`
	Goal                string    `json:"goal"`
	VersionNumber       int       `json:"version_number"`
	ValidFrom           string    `json:"valid_from"`
	ValidUntil          *string   `json:"valid_until,omitempty"`
	Status              string    `json:"status"`
	Notes               string    `json:"notes"`
	SessionCount        int       `json:"session_count"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type AssignmentDetailOutput struct {
	AssignmentSummaryOutput
	SnapshotStructure WorkoutStructureOutput `json:"snapshot_structure"`
}
