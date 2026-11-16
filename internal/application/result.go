package application

import (
	"context"
	"fmt"
	"io"

	"pos/internal/domain"

	"github.com/google/uuid"
)

type ResultService struct {
	students  StudentRepository
	forms     EvaluationFormRepository
	evaluator domain.EvaluationService
	exporter  ResultExporter
}

func NewResultService(students StudentRepository, forms EvaluationFormRepository, evaluator domain.EvaluationService, exporter ResultExporter) *ResultService {
	return &ResultService{students: students, forms: forms, evaluator: evaluator, exporter: exporter}
}

func (s *ResultService) Calculate(ctx context.Context, studentID, rubricID uuid.UUID) (domain.Score, error) {
	_, score, err := s.evaluate(ctx, studentID, rubricID)
	return score, err
}

func (s *ResultService) evaluate(ctx context.Context, studentID, rubricID uuid.UUID) (*domain.EvaluationForm, domain.Score, error) {
	form, err := s.forms.Get(ctx, studentID, rubricID)
	if err != nil {
		return nil, domain.Score{}, fmt.Errorf("get result form: %w", err)
	}
	score, err := s.evaluator.Evaluate(form)
	if err != nil {
		return nil, domain.Score{}, fmt.Errorf("calculate final score: %w", err)
	}
	return form, score, nil
}

func (s *ResultService) Export(ctx context.Context, studentID, rubricID uuid.UUID, destination io.Writer) error {
	if destination == nil {
		return fmt.Errorf("export destination must not be nil")
	}
	student, err := s.students.Get(ctx, studentID)
	if err != nil {
		return fmt.Errorf("get result student: %w", err)
	}
	form, score, err := s.evaluate(ctx, studentID, rubricID)
	if err != nil {
		return err
	}
	result := StudentResult{
		StudentID:   student.ID(),
		StudentName: student.Name(),
		RubricID:    form.RubricID(),
		RubricName:  form.Rubric().Name(),
		FinalScore:  score.Value(),
	}
	if err := s.exporter.Export(ctx, result, destination); err != nil {
		return fmt.Errorf("export student result: %w", err)
	}
	return nil
}
