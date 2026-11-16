package application

import (
	"context"
	"fmt"

	"pos/internal/domain"

	"github.com/google/uuid"
)

type RubricService struct {
	rubrics RubricRepository
}

func NewRubricService(rubrics RubricRepository) *RubricService {
	return &RubricService{rubrics: rubrics}
}

type CriterionInput struct {
	Name     string
	MaxScore float64
	Weight   int
}

// Create receives all criteria collected by the UI and saves one valid rubric.
func (s *RubricService) Create(ctx context.Context, name string, inputs []CriterionInput) (domain.Rubric, error) {
	criteria := make([]domain.Criterion, 0, len(inputs))
	for _, input := range inputs {
		maximum, err := domain.NewScore(input.MaxScore)
		if err != nil {
			return domain.Rubric{}, fmt.Errorf("create maximum score for %q: %w", input.Name, err)
		}
		criterion, err := domain.NewCriterion(input.Name, maximum, input.Weight)
		if err != nil {
			return domain.Rubric{}, fmt.Errorf("create criterion %q: %w", input.Name, err)
		}
		criteria = append(criteria, criterion)
	}
	rubric, err := domain.NewRubric(name, criteria)
	if err != nil {
		return domain.Rubric{}, fmt.Errorf("create rubric: %w", err)
	}
	if err := s.rubrics.Add(ctx, rubric); err != nil {
		return domain.Rubric{}, fmt.Errorf("save rubric: %w", err)
	}
	return rubric, nil
}

func (s *RubricService) Get(ctx context.Context, id uuid.UUID) (domain.Rubric, error) {
	rubric, err := s.rubrics.Get(ctx, id)
	if err != nil {
		return domain.Rubric{}, fmt.Errorf("get rubric: %w", err)
	}
	return rubric, nil
}

func (s *RubricService) List(ctx context.Context) ([]domain.Rubric, error) {
	rubrics, err := s.rubrics.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list rubrics: %w", err)
	}
	return rubrics, nil
}
