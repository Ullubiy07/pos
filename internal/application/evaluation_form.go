package application

import (
	"context"
	"fmt"
	"strings"

	"pos/internal/domain"

	"github.com/google/uuid"
)

type EvaluationFormService struct {
	students StudentRepository
	rubrics  RubricRepository
	forms    EvaluationFormRepository
}

func NewEvaluationFormService(students StudentRepository, rubrics RubricRepository, forms EvaluationFormRepository) *EvaluationFormService {
	return &EvaluationFormService{students: students, rubrics: rubrics, forms: forms}
}

func (s *EvaluationFormService) Create(ctx context.Context, studentID, rubricID uuid.UUID) (*domain.EvaluationForm, error) {
	if _, err := s.students.Get(ctx, studentID); err != nil {
		return nil, fmt.Errorf("get form student: %w", err)
	}
	rubric, err := s.rubrics.Get(ctx, rubricID)
	if err != nil {
		return nil, fmt.Errorf("get form rubric: %w", err)
	}
	form, err := domain.NewEvaluationForm(studentID, rubric)
	if err != nil {
		return nil, fmt.Errorf("create evaluation form: %w", err)
	}
	if err := s.forms.Create(ctx, form); err != nil {
		return nil, fmt.Errorf("save new evaluation form: %w", err)
	}
	return form, nil
}

func (s *EvaluationFormService) Get(ctx context.Context, studentID, rubricID uuid.UUID) (*domain.EvaluationForm, error) {
	form, err := s.forms.Get(ctx, studentID, rubricID)
	if err != nil {
		return nil, fmt.Errorf("get evaluation form: %w", err)
	}
	if form == nil {
		return nil, fmt.Errorf("get evaluation form: %w", ErrNotFound)
	}
	return form, nil
}

func (s *EvaluationFormService) AddEvaluation(ctx context.Context, studentID, rubricID uuid.UUID, criterionName string, score float64, comment string) error {
	return s.change(ctx, studentID, rubricID, func(form *domain.EvaluationForm) error {
		evaluation, err := evaluationFor(form, criterionName, score, comment)
		if err != nil {
			return err
		}
		return form.AddEvaluation(evaluation)
	})
}

func (s *EvaluationFormService) UpdateEvaluation(ctx context.Context, studentID, rubricID uuid.UUID, criterionName string, score float64, comment string) error {
	return s.change(ctx, studentID, rubricID, func(form *domain.EvaluationForm) error {
		evaluation, err := evaluationFor(form, criterionName, score, comment)
		if err != nil {
			return err
		}
		return form.UpdateEvaluation(evaluation)
	})
}

func (s *EvaluationFormService) RemoveEvaluation(ctx context.Context, studentID, rubricID uuid.UUID, criterionName string) error {
	return s.change(ctx, studentID, rubricID, func(form *domain.EvaluationForm) error {
		return form.RemoveEvaluation(strings.TrimSpace(criterionName))
	})
}

func (s *EvaluationFormService) Complete(ctx context.Context, studentID, rubricID uuid.UUID) error {
	return s.change(ctx, studentID, rubricID, func(form *domain.EvaluationForm) error {
		return form.Complete()
	})
}

func (s *EvaluationFormService) Reopen(ctx context.Context, studentID, rubricID uuid.UUID) error {
	return s.change(ctx, studentID, rubricID, func(form *domain.EvaluationForm) error {
		return form.Reopen()
	})
}

func (s *EvaluationFormService) change(ctx context.Context, studentID, rubricID uuid.UUID, change func(*domain.EvaluationForm) error) error {
	form, err := s.Get(ctx, studentID, rubricID)
	if err != nil {
		return err
	}
	if err := change(form); err != nil {
		return fmt.Errorf("change evaluation form: %w", err)
	}
	if err := s.forms.Save(ctx, form); err != nil {
		return fmt.Errorf("save evaluation form: %w", err)
	}
	return nil
}

func evaluationFor(form *domain.EvaluationForm, criterionName string, value float64, comment string) (domain.Evaluation, error) {
	criterionName = strings.TrimSpace(criterionName)
	for _, criterion := range form.Rubric().Criteria() {
		if criterion.Name() != criterionName {
			continue
		}
		score, err := domain.NewScore(value)
		if err != nil {
			return domain.Evaluation{}, fmt.Errorf("create evaluation score: %w", err)
		}
		evaluation, err := domain.NewEvaluation(criterion, score)
		if err != nil {
			return domain.Evaluation{}, fmt.Errorf("create evaluation: %w", err)
		}
		evaluation.Comment = comment
		return *evaluation, nil
	}
	return domain.Evaluation{}, fmt.Errorf("criterion %q in form rubric: %w", criterionName, ErrNotFound)
}
