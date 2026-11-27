package workout

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type WorkoutPlan struct {
	ID          uuid.UUID
	TrainerID   uuid.UUID
	StudentID   uuid.UUID
	Title       string
	Description string
	IsActive    bool
	Sessions    []WorkoutSession
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type WorkoutSession struct {
	ID            uuid.UUID
	WorkoutPlanID uuid.UUID
	Title         string
	DayOfWeek     int
	Notes         string
	Exercises     []Exercise
	CreatedAt     time.Time
}

type Exercise struct {
	ID                uuid.UUID
	WorkoutSessionID  uuid.UUID
	CatalogExerciseID *uuid.UUID
	Name              string
	Sets              int
	Reps              string
	RestSeconds       int
	Notes             string
	VideoURL          string
	Position          int
	CreatedAt         time.Time
}

func NewWorkoutPlan(trainerID, studentID uuid.UUID, title, description string) (*WorkoutPlan, error) {
	if trainerID == uuid.Nil {
		return nil, errors.New("o ID do treinador é obrigatório")
	}
	if studentID == uuid.Nil {
		return nil, errors.New("o ID do aluno é obrigatório")
	}
	if len(title) < 3 {
		return nil, errors.New("o título deve ter ao menos 3 caracteres")
	}
	now := time.Now()
	return &WorkoutPlan{
		ID:          uuid.New(),
		TrainerID:   trainerID,
		StudentID:   studentID,
		Title:       title,
		Description: description,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func NewWorkoutSession(planID uuid.UUID, title string, dayOfWeek int) (*WorkoutSession, error) {
	if planID == uuid.Nil {
		return nil, errors.New("o ID do plano é obrigatório")
	}
	if len(title) < 2 {
		return nil, errors.New("o título da sessão deve ter ao menos 2 caracteres")
	}
	if dayOfWeek < 0 || dayOfWeek > 6 {
		return nil, errors.New("dia da semana deve ser entre 0 (domingo) e 6 (sábado)")
	}
	return &WorkoutSession{
		ID:            uuid.New(),
		WorkoutPlanID: planID,
		Title:         title,
		DayOfWeek:     dayOfWeek,
		CreatedAt:     time.Now(),
	}, nil
}

func NewExercise(sessionID uuid.UUID, name string, sets int, reps string, restSeconds, position int) (*Exercise, error) {
	if sessionID == uuid.Nil {
		return nil, errors.New("o ID da sessão é obrigatório")
	}
	if len(name) < 2 {
		return nil, errors.New("o nome do exercício deve ter ao menos 2 caracteres")
	}
	if sets < 0 {
		return nil, errors.New("número de séries não pode ser negativo")
	}
	return &Exercise{
		ID:               uuid.New(),
		WorkoutSessionID: sessionID,
		Name:             name,
		Sets:             sets,
		Reps:             reps,
		RestSeconds:      restSeconds,
		Position:         position,
		CreatedAt:        time.Now(),
	}, nil
}
