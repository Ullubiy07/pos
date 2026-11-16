package domain

import (
	"fmt"
	"slices"
)

type Rubric struct {
	items []Criterion
}

func NewRubric(criteria []Criterion) (Rubric, error) {
	cMap := make(map[string]Criterion, len(criteria))
	items := []Criterion{}

	weight := 0
	for _, criterion := range criteria {
		if err := criterion.validate(); err != nil {
			return Rubric{}, fmt.Errorf("invalid rubric criterion: %w", err)
		}
		if _, exist := cMap[criterion.name]; exist {
			return Rubric{}, fmt.Errorf("found criterion duplicate")
		}
		weight += criterion.weight
		cMap[criterion.name] = criterion
		items = append(items, criterion)
	}

	if weight != 100 {
		return Rubric{}, fmt.Errorf("total weight of criteria must amount to 100 percent")
	}
	return Rubric{items: items}, nil
}

func (r Rubric) Contains(criterion Criterion) bool {
	return slices.Contains(r.items, criterion)
}

func (r Rubric) Criteria() []Criterion {
	return slices.Clone(r.items)
}
