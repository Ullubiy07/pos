package domain

import (
	"fmt"
	"math"
	"strings"
)

type Score struct {
	value float64
}

func NewScore(value float64) (Score, error) {
	if math.IsInf(value, 0) || math.IsNaN(value) || value < 0 {
		return Score{}, fmt.Errorf("invalid score")
	}
	return Score{value: value}, nil
}

func (s Score) Value() float64 { return s.value }

type Criterion struct {
	name     string
	maxScore Score
	weight   int
}

func NewCriterion(name string, maxScore Score, weight int) (Criterion, error) {
	criterion := Criterion{name: strings.TrimSpace(name), maxScore: maxScore, weight: weight}
	if err := criterion.validate(); err != nil {
		return Criterion{}, err
	}
	return criterion, nil
}

func (c Criterion) Name() string { return c.name }

func (c Criterion) MaxScore() Score { return c.maxScore }

func (c Criterion) Weight() int { return c.weight }

func (c Criterion) validate() error {
	if c.weight < 0 || c.weight > 100 {
		return fmt.Errorf("weight must be between 0 and 100 percent")
	}
	if _, err := NewScore(c.maxScore.value); err != nil {
		return fmt.Errorf("invalid maximum score: %w", err)
	}
	if c.maxScore.value == 0 {
		return fmt.Errorf("max score must be positive")
	}
	if strings.TrimSpace(c.name) == "" {
		return fmt.Errorf("name must be non-empty")
	}
	return nil
}
