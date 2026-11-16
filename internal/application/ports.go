package application

import (
	"context"
	"errors"
	"io"

	"pos/internal/domain"

	"github.com/google/uuid"
)

// Repository implementations must preserve these errors for errors.Is checks.
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type StudentRepository interface {
	Add(context.Context, domain.Student) error
	Get(context.Context, uuid.UUID) (domain.Student, error)
	List(context.Context) ([]domain.Student, error)
}

// Rubrics are stored as a whole, including all of their criteria.
// Add rejects an existing ID; criterion names are unique within each rubric.
type RubricRepository interface {
	Add(context.Context, domain.Rubric) error
	Get(context.Context, uuid.UUID) (domain.Rubric, error)
	List(context.Context) ([]domain.Rubric, error)
}

// A form is identified by the pair (StudentID, RubricID).
// Reads and writes must use detached snapshots: changing a returned form must
// not change stored state until Save succeeds.
type EvaluationFormRepository interface {
	// Create rejects an existing (StudentID, RubricID) pair with ErrAlreadyExists.
	Create(context.Context, *domain.EvaluationForm) error
	Get(ctx context.Context, studentID, rubricID uuid.UUID) (*domain.EvaluationForm, error)
	Save(context.Context, *domain.EvaluationForm) error
}

// StudentResult is export data, independent of CSV, JSON, or console formatting.
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
