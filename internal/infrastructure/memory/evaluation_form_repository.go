package memory

import (
	"context"
	"fmt"
	"sync"

	"pos/internal/application"
	"pos/internal/domain"

	"github.com/google/uuid"
)

type formKey struct {
	studentID uuid.UUID
	rubricID  uuid.UUID
}

type EvaluationFormRepository struct {
	mu    sync.RWMutex
	items map[formKey]*domain.EvaluationForm
}

var _ application.EvaluationFormRepository = (*EvaluationFormRepository)(nil)

func NewEvaluationFormRepository() *EvaluationFormRepository {
	return &EvaluationFormRepository{items: make(map[formKey]*domain.EvaluationForm)}
}

func (r *EvaluationFormRepository) Create(ctx context.Context, form *domain.EvaluationForm) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	copy, err := cloneForm(form)
	if err != nil {
		return fmt.Errorf("copy new form: %w", err)
	}
	key := formKey{copy.StudentID(), copy.RubricID()}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[key]; exists {
		return fmt.Errorf("form for student %s and rubric %s: %w", key.studentID, key.rubricID, application.ErrAlreadyExists)
	}
	if r.items == nil {
		r.items = make(map[formKey]*domain.EvaluationForm)
	}
	r.items[key] = copy
	return nil
}

func (r *EvaluationFormRepository) Get(ctx context.Context, studentID, rubricID uuid.UUID) (*domain.EvaluationForm, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	form, exists := r.items[formKey{studentID, rubricID}]
	if !exists {
		return nil, fmt.Errorf("form for student %s and rubric %s: %w", studentID, rubricID, application.ErrNotFound)
	}
	return cloneForm(form)
}

func (r *EvaluationFormRepository) Save(ctx context.Context, form *domain.EvaluationForm) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	copy, err := cloneForm(form)
	if err != nil {
		return fmt.Errorf("copy updated form: %w", err)
	}
	key := formKey{copy.StudentID(), copy.RubricID()}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[key]; !exists {
		return fmt.Errorf("form for student %s and rubric %s: %w", key.studentID, key.rubricID, application.ErrNotFound)
	}
	r.items[key] = copy
	return nil
}

func cloneForm(form *domain.EvaluationForm) (*domain.EvaluationForm, error) {
	if form == nil {
		return nil, fmt.Errorf("evaluation form must not be nil")
	}
	copy, err := domain.NewEvaluationForm(form.StudentID(), form.Rubric())
	if err != nil {
		return nil, err
	}
	for _, evaluation := range form.Evaluations() {
		if err := copy.AddEvaluation(evaluation); err != nil {
			return nil, err
		}
	}
	switch form.Status() {
	case domain.Draft:
	case domain.Completed:
		if err := copy.Complete(); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown evaluation form status %q", form.Status())
	}
	return copy, nil
}
