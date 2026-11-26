package student

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

type Student struct {
	ID           uuid.UUID  `db:"id"`
	UserID       uuid.UUID  `db:"user_id"`
	TrainerID    *uuid.UUID `db:"trainer_id"`
	CPFEncrypted *string    `db:"cpf_encrypted"`
	CPFHash      *string    `db:"cpf_hash"`
	CreatedAt    time.Time  `db:"created_at"`

	cleanCPF string
}

type StudentBaseInfo struct {
	ID       string `db:"id" json:"id"`
	FullName string `db:"full_name" json:"full_name"`
	Email    string `db:"email" json:"email"`
}

var numericOnlyRegex = regexp.MustCompile(`[^0-9]+`)

func NewStudent(userID, trainerID uuid.UUID, rawCPF string) (*Student, error) {
	if userID == uuid.Nil {
		return nil, errors.New("o UserID é obrigatório")
	}
	if trainerID == uuid.Nil {
		return nil, errors.New("o TrainerID é obrigatório")
	}

	cleanCPF := numericOnlyRegex.ReplaceAllString(rawCPF, "")
	if cleanCPF != "" && len(cleanCPF) != 11 {
		return nil, errors.New("o CPF deve conter exatamente 11 digitos numericos")
	}

	tid := trainerID
	return &Student{
		ID:        uuid.New(),
		UserID:    userID,
		TrainerID: &tid,
		cleanCPF:  cleanCPF,
		CreatedAt: time.Now(),
	}, nil
}

func (s *Student) GetCleanCPF() string {
	return s.cleanCPF
}

func (s *Student) ClearSensitiveData() {
	s.cleanCPF = ""
}

type BiometricData struct {
	ID                uuid.UUID `db:"id"`
	StudentID         uuid.UUID `db:"student_id"`
	WeightKg          float64   `db:"weight_kg"`
	HeightCm          float64   `db:"height_cm"`
	BMI               float64   `db:"bmi"`
	BodyFatPercentage float64   `db:"body_fat_percentage"`
	RecordedBy        uuid.UUID `db:"recorded_by"`
	RecordedAt        time.Time `db:"recorded_at"`
}

func NewBiometricRecord(studentID, trainerID uuid.UUID, weightKg, heightCm, fat float64) (*BiometricData, error) {
	if weightKg <= 0 || heightCm <= 0 {
		return nil, errors.New("peso e altura precisam ser maiores que zero")
	}

	heightInMeters := heightCm / 100.0
	bmi := weightKg / (heightInMeters * heightInMeters)

	return &BiometricData{
		ID:                uuid.New(),
		StudentID:         studentID,
		WeightKg:          weightKg,
		HeightCm:          heightCm,
		BMI:               bmi,
		BodyFatPercentage: fat,
		RecordedBy:        trainerID,
		RecordedAt:        time.Now(),
	}, nil
}
