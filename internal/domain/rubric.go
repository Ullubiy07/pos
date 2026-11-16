package domain

import (
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
)

type Rubric struct {
	id    uuid.UUID
	name  string
	items []Criterion
}

func NewRubric(name string, criteria []Criterion) (Rubric, error) {
	rubric := Rubric{id: uuid.New(), name: strings.TrimSpace(name), items: slices.Clone(criteria)}
	if err := rubric.validate(); err != nil {
		return Rubric{}, err
	}
	return rubric, nil
}

func (r Rubric) ID() uuid.UUID { return r.id }

func (r Rubric) Name() string { return r.name }

func (r Rubric) validate() error {
	if r.id == uuid.Nil {
		return fmt.Errorf("rubric ID must not be empty")
	}
	if strings.TrimSpace(r.name) == "" {
		return fmt.Errorf("rubric name must not be empty")
	}
	names := make(map[string]struct{}, len(r.items))

	weight := 0
	for _, criterion := range r.items {
		if err := criterion.validate(); err != nil {
			return fmt.Errorf("invalid rubric criterion: %w", err)
		}
		if _, exist := names[criterion.name]; exist {
			return fmt.Errorf("found criterion duplicate")
		}
		weight += criterion.weight
		names[criterion.name] = struct{}{}
	}

	if weight != 100 {
		return fmt.Errorf("total weight of criteria must amount to 100 percent")
	}
	return nil
}

func (r Rubric) Contains(criterion Criterion) bool {
	return slices.Contains(r.items, criterion)
}

func (r Rubric) Criteria() []Criterion {
	return slices.Clone(r.items)
}
