package domain

import (
	"math"
	"testing"

	"github.com/google/uuid"
)

func TestEvaluationServiceUsesScorePolicy(t *testing.T) {
	maximum, err := NewScore(3)
	if err != nil {
		t.Fatal(err)
	}
	criterion, err := NewCriterion("Exam", maximum, 100)
	if err != nil {
		t.Fatal(err)
	}
	rubric, err := NewRubric("Course", []Criterion{criterion})
	if err != nil {
		t.Fatal(err)
	}
	form, err := NewEvaluationForm(uuid.New(), rubric)
	if err != nil {
		t.Fatal(err)
	}
	score, err := NewScore(1)
	if err != nil {
		t.Fatal(err)
	}
	evaluation, err := NewEvaluation(criterion, score)
	if err != nil {
		t.Fatal(err)
	}
	if err := form.AddEvaluation(*evaluation); err != nil {
		t.Fatal(err)
	}
	if err := form.Complete(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		policy ScorePolicy
		want   float64
	}{
		{ExactScorePolicy{}, 100.0 / 3},
		{HundredthsScorePolicy{}, 33.33},
	} {
		result, err := NewEvaluationService(tc.policy).Evaluate(form)
		if err != nil || math.Abs(result.Value()-tc.want) > 1e-10 {
			t.Fatalf("policy %T: got %v, %v; want %v", tc.policy, result.Value(), err, tc.want)
		}
	}
	if _, err := NewEvaluationService(nil).Evaluate(form); err == nil {
		t.Fatal("nil policy accepted")
	}
}
