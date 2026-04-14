package student

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// Student representa o domínio de um Aluno na nossa aplicação.
// Aqui na camada de domínio nós retemos o "cleanCPF" em estado transitório
// apenas para trafegar entre HTTP <-> UseCase antes dele ser mascarado e "apagado" com AES.
type Student struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	CPFEncrypted string
	CPFHash      string
	CreatedAt    time.Time

	// Transitório - a camada de infraestrutura/db não deve ver isso
	cleanCPF string
}

// Regex para detectar digitos numéricos
var numericOnlyRegex = regexp.MustCompile(`[^0-9]+`)

// NewStudent cria uma nova instancia validando regras imutaveis.
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

// GetCleanCPF expõe a string que será processada pela camada de UseCases (AES/HMAC).
func (s *Student) GetCleanCPF() string {
	return s.cleanCPF
}

// ClearSensitiveData é o mecanismo para limpar da memória o CPF após o seu uso.
func (s *Student) ClearSensitiveData() {
	s.cleanCPF = ""
}

// BiometricData representa uma entrada no histórico fisiológico do aluno
type BiometricData struct {
	ID                uuid.UUID
	StudentID         uuid.UUID
	WeightKg          float64
	HeightCm          float64
	BMI               float64
	BodyFatPercentage float64
	// MuscleMassKg, NotesEncrypted...
	RecordedBy uuid.UUID
	RecordedAt time.Time
}

// NewBiometricRecord constrói um registro garantindo leis da termodinamica básicas e calculos.
func NewBiometricRecord(studentID, trainerID uuid.UUID, weightKg, heightCm, fat float64) (*BiometricData, error) {
	if weightKg <= 0 || heightCm <= 0 {
		return nil, errors.New("peso e altura precisam ser maiores que zero")
	}

	// ICM (BMI) = Massa (Kg) / (Altura(M) * Altura(M))
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
