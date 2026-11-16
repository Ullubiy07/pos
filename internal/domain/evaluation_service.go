package domain

import (
	"fmt"
)

type EvaluationService struct{}

func NewEvaluationService() EvaluationService {
	return EvaluationService{}
}

func (s EvaluationService) Evaluate(form *EvaluationForm) (Score, error) {
	if form == nil {
		return Score{}, fmt.Errorf("evaluation form must not be nil")
	}
	if form.Status() != Completed {
		return Score{}, fmt.Errorf("only a completed form can be evaluated")
	}

	var total float64
	for _, evaluation := range form.evaluations {
		total += evaluation.contribution.value
	}
	score, err := NewScore(total)
	if err != nil {
		return Score{}, fmt.Errorf("calculate final score: %w", err)
	}
	return score, nil
}
