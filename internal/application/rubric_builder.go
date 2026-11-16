package application

import (
	"fmt"
	"strings"

	"pos/internal/domain"
)

// Ошибка при добавлении критерия не сбрасывает уже введённые данные.
type RubricBuilder struct {
	name     string
	criteria []domain.Criterion
	weight   int
}

func NewRubricBuilder(name string) (*RubricBuilder, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("rubric name must not be empty")
	}
	return &RubricBuilder{name: name}, nil
}

func (b *RubricBuilder) AddCriterion(input CriterionInput) error {
	maximum, err := domain.NewScore(input.MaxScore)
	if err != nil {
		return fmt.Errorf("create maximum score for %q: %w", input.Name, err)
	}
	criterion, err := domain.NewCriterion(input.Name, maximum, input.Weight)
	if err != nil {
		return fmt.Errorf("create criterion %q: %w", input.Name, err)
	}
	for _, existing := range b.criteria {
		if existing.Name() == criterion.Name() {
			return fmt.Errorf("criterion %q already added", criterion.Name())
		}
	}
	if b.weight+criterion.Weight() > 100 {
		return fmt.Errorf("total criterion weight would exceed 100 percent")
	}
	b.criteria = append(b.criteria, criterion)
	b.weight += criterion.Weight()
	return nil
}

func (b *RubricBuilder) Build() (domain.Rubric, error) {
	return domain.NewRubric(b.name, b.criteria)
}
