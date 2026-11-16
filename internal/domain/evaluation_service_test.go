package domain

import "testing"

func TestEvaluationServiceRequiresCompletedForm(t *testing.T) {
	service := NewEvaluationService()
	if _, err := service.Evaluate(nil); err == nil {
		t.Fatal("nil form accepted")
	}
	form, first, second := formFixture(t)
	if _, err := service.Evaluate(form); err == nil {
		t.Fatal("draft form accepted")
	}
	if err := form.AddEvaluation(first); err != nil {
		t.Fatal(err)
	}
	if err := form.AddEvaluation(second); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Evaluate(form); err == nil {
		t.Fatal("fully filled draft accepted before completion")
	}
}

func TestEvaluationServiceUsesCurrentFormEvaluations(t *testing.T) {
	service := NewEvaluationService()
	form, first, second := formFixture(t)
	if err := form.AddEvaluation(first); err != nil {
		t.Fatal(err)
	}
	if err := form.AddEvaluation(second); err != nil {
		t.Fatal(err)
	}
	if err := form.Complete(); err != nil {
		t.Fatal(err)
	}
	result, err := service.Evaluate(form)
	if err != nil || result.value != 100 {
		t.Fatalf("full scores: got %v, %v; want 100", result.value, err)
	}
	if err := form.Reopen(); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Evaluate(form); err == nil {
		t.Fatal("reopened form accepted")
	}
	score, err := NewScore(5)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := NewEvaluation(first.criterion, score)
	if err != nil {
		t.Fatal(err)
	}
	if err := form.UpdateEvaluation(*updated); err != nil {
		t.Fatal(err)
	}
	if err := form.Complete(); err != nil {
		t.Fatal(err)
	}
	result, err = service.Evaluate(form)
	if err != nil || result.value != 70 {
		t.Fatalf("updated scores: got %v, %v; want 70", result.value, err)
	}
}
