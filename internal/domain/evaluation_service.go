package domain

import (
	"fmt"
)

type EvaluationService struct {
	policy ScorePolicy
}

func NewEvaluationService(policy ScorePolicy) EvaluationService {
	return EvaluationService{policy: policy}
}

func (s EvaluationService) Evaluate(form *EvaluationForm) (Score, error) {
	if form == nil {
		return Score{}, fmt.Errorf("evaluation form must not be nil")
	}
	if form.Status() != Completed {
		return Score{}, fmt.Errorf("only a completed form can be evaluated")
	}
	if s.policy == nil {
		return Score{}, fmt.Errorf("final score policy must not be nil")
	}

	var total float64
	for _, evaluation := range form.evaluations {
		total += evaluation.contribution.value
	}
	score, err := NewScore(total)
	if err != nil {
		return Score{}, fmt.Errorf("calculate final score: %w", err)
	}
	result, err := s.policy.Apply(score)
	if err != nil {
		return Score{}, fmt.Errorf("apply final score policy: %w", err)
	}
	return result, nil
}
