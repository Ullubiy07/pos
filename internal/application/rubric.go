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

func (s *RubricService) Create(ctx context.Context, name string, inputs []CriterionInput) (domain.Rubric, error) {
	builder, err := NewRubricBuilder(name)
	if err != nil {
		return domain.Rubric{}, err
	}
	for _, input := range inputs {
		if err := builder.AddCriterion(input); err != nil {
			return domain.Rubric{}, err
		}
	}
	return s.Save(ctx, builder)
}

func (s *RubricService) Save(ctx context.Context, builder *RubricBuilder) (domain.Rubric, error) {
	if builder == nil {
		return domain.Rubric{}, fmt.Errorf("rubric builder must not be nil")
	}
	rubric, err := builder.Build()
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
