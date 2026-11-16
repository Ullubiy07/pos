package domain

import (
	"fmt"
	"time"
)

type Evaluation struct {
	Comment      string
	criterion    Criterion
	score        Score
	contribution Score
	createdAt    time.Time
}

func NewEvaluation(criterion Criterion, score Score) (*Evaluation, error) {
	if err := criterion.validate(); err != nil {
		return nil, fmt.Errorf("invalid evaluation criterion: %w", err)
	}
	if _, err := NewScore(score.value); err != nil {
		return nil, fmt.Errorf("invalid evaluation score: %w", err)
	}
	if score.value > criterion.maxScore.value {
		return nil, fmt.Errorf("score value exceeded max score for criterion")
	}

	contribution, err := NewScore(score.value / criterion.maxScore.value * float64(criterion.weight))
	if err != nil {
		return nil, fmt.Errorf("calculate evaluation contribution: %w", err)
	}

	return &Evaluation{
		criterion: criterion,
		score:     score, createdAt: time.Now(),
		contribution: contribution,
	}, nil
}

func (e Evaluation) Criterion() Criterion { return e.criterion }

func (e Evaluation) Score() Score { return e.score }

func (e Evaluation) Contribution() Score { return e.contribution }

func (e Evaluation) CreatedAt() time.Time { return e.createdAt }
