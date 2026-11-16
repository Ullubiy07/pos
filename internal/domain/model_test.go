package domain_test

import (
	"math"
	"testing"

	"pos/internal/domain"

	"github.com/google/uuid"
)

func TestScoreRejectsInvalidNumbers(t *testing.T) {
	for _, value := range []float64{-1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := domain.NewScore(value); err == nil {
			t.Errorf("accepted invalid score %v", value)
		}
	}
	if score, err := domain.NewScore(0); err != nil || score.Value() != 0 {
		t.Fatalf("zero score must be valid: %v", err)
	}
}

func TestStudentName(t *testing.T) {
	if _, err := domain.NewStudent(" \t\n"); err == nil {
		t.Fatal("blank name accepted")
	}
	student, err := domain.NewStudent(" Alice ")
	if err != nil {
		t.Fatal(err)
	}
	if student.Name() != "Alice" || student.ID() == uuid.Nil {
		t.Fatal("expected trimmed name and generated ID")
	}
}

func TestCriterionAndRubricInvariants(t *testing.T) {
	max, err := domain.NewScore(10)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		max    domain.Score
		weight int
	}{
		{" \t", max, 100},
		{"exam", domain.Score{}, 100},
		{"exam", max, -1},
		{"exam", max, 101},
	} {
		if _, err := domain.NewCriterion(tc.name, tc.max, tc.weight); err == nil {
			t.Errorf("invalid criterion accepted: %+v", tc)
		}
	}
	criterion, err := domain.NewCriterion(" exam ", max, 50)
	if err != nil {
		t.Fatal(err)
	}
	if criterion.Name() != "exam" || criterion.Weight() != 50 || criterion.MaxScore() != max {
		t.Fatal("unexpected criterion properties")
	}
	if _, err := domain.NewRubric("Course", []domain.Criterion{criterion, criterion}); err == nil {
		t.Fatal("duplicate criteria accepted")
	}
	if _, err := domain.NewRubric("Course", []domain.Criterion{criterion}); err == nil {
		t.Fatal("weights not totaling 100 accepted")
	}
	if _, err := domain.NewRubric("Course", []domain.Criterion{{}}); err == nil {
		t.Fatal("zero-value criterion accepted")
	}
	full, err := domain.NewCriterion("exam", max, 100)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := domain.NewRubric(" \t", []domain.Criterion{full}); err == nil {
		t.Fatal("blank rubric name accepted")
	}
}

func TestFormValidationAndRubricIsolation(t *testing.T) {
	max, err := domain.NewScore(10)
	if err != nil {
		t.Fatal(err)
	}
	criterion, err := domain.NewCriterion("exam", max, 100)
	if err != nil {
		t.Fatal(err)
	}
	input := []domain.Criterion{criterion}
	rubric, err := domain.NewRubric("Course", input)
	if err != nil {
		t.Fatal(err)
	}
	input[0] = domain.Criterion{}
	output := rubric.Criteria()
	output[0] = domain.Criterion{}
	if !rubric.Contains(criterion) {
		t.Fatal("external slice mutation changed rubric")
	}
	if _, err := domain.NewEvaluationForm(uuid.Nil, rubric); err == nil {
		t.Fatal("empty student ID accepted")
	}
	if _, err := domain.NewEvaluationForm(uuid.New(), domain.Rubric{}); err == nil {
		t.Fatal("empty rubric accepted")
	}
	form, err := domain.NewEvaluationForm(uuid.New(), rubric)
	if err != nil {
		t.Fatal(err)
	}
	criteria := form.Rubric().Criteria()
	criteria[0] = domain.Criterion{}
	if !form.Rubric().Contains(criterion) {
		t.Fatal("external mutation changed form rubric")
	}
	if err := form.AddEvaluation(domain.Evaluation{}); err == nil {
		t.Fatal("zero-value evaluation accepted")
	}
	tooHigh, err := domain.NewScore(11)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := domain.NewEvaluation(criterion, tooHigh); err == nil {
		t.Fatal("score above maximum accepted")
	}
	if _, err := domain.NewEvaluation(domain.Criterion{}, domain.Score{}); err == nil {
		t.Fatal("zero-value criterion accepted in evaluation")
	}
}

func TestFinalScoreNormalizesDifferentMaxima(t *testing.T) {
	var criteria []domain.Criterion
	var evaluations []domain.Evaluation
	for _, tc := range []struct {
		name  string
		max   float64
		score float64
	}{
		{"exam", 100, 80},
		{"project", 10, 5},
	} {
		max, err := domain.NewScore(tc.max)
		if err != nil {
			t.Fatal(err)
		}
		criterion, err := domain.NewCriterion(tc.name, max, 50)
		if err != nil {
			t.Fatal(err)
		}
		score, err := domain.NewScore(tc.score)
		if err != nil {
			t.Fatal(err)
		}
		evaluation, err := domain.NewEvaluation(criterion, score)
		if err != nil {
			t.Fatal(err)
		}
		if evaluation.Criterion() != criterion || evaluation.Score() != score || evaluation.CreatedAt().IsZero() {
			t.Fatal("unexpected evaluation properties")
		}
		criteria = append(criteria, criterion)
		evaluations = append(evaluations, *evaluation)
	}
	rubric, err := domain.NewRubric("Course", criteria)
	if err != nil {
		t.Fatal(err)
	}
	form, err := domain.NewEvaluationForm(uuid.New(), rubric)
	if err != nil {
		t.Fatal(err)
	}
	for _, evaluation := range evaluations {
		if err := form.AddEvaluation(evaluation); err != nil {
			t.Fatal(err)
		}
	}
	if err := form.Complete(); err != nil {
		t.Fatal(err)
	}
	result, err := domain.NewEvaluationService(domain.ExactScorePolicy{}).Evaluate(form)
	if err != nil || math.Abs(result.Value()-65) > 1e-9 {
		t.Fatalf("got %v, %v; want 65", result.Value(), err)
	}
}
