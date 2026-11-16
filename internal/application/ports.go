package application

import (
	"context"
	"errors"
	"io"

	"pos/internal/domain"

	"github.com/google/uuid"
)

// Адаптеры возвращают эти ошибки напрямую или оборачивают через %w.
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type StudentRepository interface {
	Add(context.Context, domain.Student) error
	Get(context.Context, uuid.UUID) (domain.Student, error)
	List(context.Context) ([]domain.Student, error)
}

// Рубрикатор хранится целиком с критериями. Повторный ID — ErrAlreadyExists.
type RubricRepository interface {
	Add(context.Context, domain.Rubric) error
	Get(context.Context, uuid.UUID) (domain.Rubric, error)
	List(context.Context) ([]domain.Rubric, error)
}

// Ключ формы — пара StudentID + RubricID.
// Get возвращает независимую копию; изменения сохраняются только через Save.
type EvaluationFormRepository interface {
	// Повторная пара StudentID + RubricID — ErrAlreadyExists.
	Create(context.Context, *domain.EvaluationForm) error
	Get(ctx context.Context, studentID, rubricID uuid.UUID) (*domain.EvaluationForm, error)
	Save(context.Context, *domain.EvaluationForm) error
}

type StudentResult struct {
	StudentID   uuid.UUID
	StudentName string
	RubricID    uuid.UUID
	RubricName  string
	FinalScore  float64
}

type ResultExporter interface {
	Export(context.Context, StudentResult, io.Writer) error
}
