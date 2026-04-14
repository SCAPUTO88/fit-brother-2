package student

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

type Student struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	CPFEncrypted string
	CPFHash      string
	CreatedAt    time.Time

	cleanCPF string
}

var numericOnlyRegex = regexp.MustCompile(`[^0-9]+`)

func NewStudent(userID uuid.UUID, rawCPF string) (*Student, error) {
	if userID == uuid.Nil {
		return nil, errors.New("o UserID é obrigatório")
	}

	cleanCPF := numericOnlyRegex.ReplaceAllString(rawCPF, "")
	if len(cleanCPF) != 11 {
		return nil, errors.New("o CPF deve conter exatamente 11 digitos numericos")
	}

	return &Student{
		ID:        uuid.New(),
		UserID:    userID,
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
	ID                uuid.UUID
	StudentID         uuid.UUID
	WeightKg          float64
	HeightCm          float64
	BMI               float64
	BodyFatPercentage float64
	RecordedBy        uuid.UUID
	RecordedAt        time.Time
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
