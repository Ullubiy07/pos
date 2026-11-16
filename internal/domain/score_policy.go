package domain

import "math"

type ScorePolicy interface {
	Apply(Score) (Score, error)
}

type ExactScorePolicy struct{}

func (ExactScorePolicy) Apply(score Score) (Score, error) {
	return NewScore(score.Value())
}

type HundredthsScorePolicy struct{}

func (HundredthsScorePolicy) Apply(score Score) (Score, error) {
	return NewScore(math.Round(score.Value()*100) / 100)
}
