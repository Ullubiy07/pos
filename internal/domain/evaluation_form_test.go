package domain

import (
	"testing"

	"github.com/google/uuid"
)

func formFixture(t *testing.T) (*EvaluationForm, Evaluation, Evaluation) {
	t.Helper()
	maxScore, err := NewScore(10)
	if err != nil {
		t.Fatal(err)
	}
	first, err := NewCriterion("exam", maxScore, 60)
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewCriterion("project", maxScore, 40)
	if err != nil {
		t.Fatal(err)
	}
	rubric, err := NewRubric([]Criterion{first, second})
	if err != nil {
		t.Fatal(err)
	}
	a, err := NewEvaluation(first, maxScore)
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewEvaluation(second, maxScore)
	if err != nil {
		t.Fatal(err)
	}
	form, err := NewEvaluationForm(uuid.New(), rubric)
	if err != nil {
		t.Fatal(err)
	}
	return form, *a, *b
}

func TestEvaluationFormLifecycle(t *testing.T) {
	f, first, second := formFixture(t)
	if f.Status() != Draft || f.StudentID() == uuid.Nil {
		t.Fatal("new form must be a student-bound draft")
	}
	if err := f.Complete(); err == nil {
		t.Fatal("incomplete form was completed")
	}
	if err := f.AddEvaluation(first); err != nil {
		t.Fatal(err)
	}
	if err := f.AddEvaluation(second); err != nil {
		t.Fatal(err)
	}
	if err := f.Complete(); err != nil {
		t.Fatal(err)
	}
	if f.Status() != Completed {
		t.Fatal("form must be completed")
	}
	for _, change := range []func() error{
		func() error { return f.AddEvaluation(first) },
		func() error { return f.UpdateEvaluation(first) },
		func() error { return f.RemoveEvaluation(first.criterion.name) },
		f.Complete,
	} {
		if err := change(); err == nil {
			t.Fatal("completed form accepted a change")
		}
	}
	if err := f.Reopen(); err != nil {
		t.Fatal(err)
	}
	if f.Status() != Draft {
		t.Fatal("reopened form must be a draft")
	}
	if err := f.RemoveEvaluation(second.criterion.name); err != nil {
		t.Fatal(err)
	}
	if err := f.Complete(); err == nil {
		t.Fatal("form with a removed evaluation was completed")
	}
}

func TestEvaluationFormRejectsDuplicatesAndForeignCriteria(t *testing.T) {
	f, first, _ := formFixture(t)
	if err := f.AddEvaluation(first); err != nil {
		t.Fatal(err)
	}
	if err := f.AddEvaluation(first); err == nil {
		t.Fatal("duplicate evaluation accepted")
	}
	foreign := first
	foreign.criterion.weight++
	if err := f.AddEvaluation(foreign); err == nil {
		t.Fatal("same name with a different weight accepted")
	}
	if err := f.UpdateEvaluation(foreign); err == nil {
		t.Fatal("update with a foreign criterion accepted")
	}
	if err := f.Complete(); err == nil {
		t.Fatal("duplicate attempt must not fill another criterion")
	}
}

func TestEvaluationFormUpdateAndReadIsolation(t *testing.T) {
	f, first, second := formFixture(t)
	if err := f.UpdateEvaluation(first); err == nil {
		t.Fatal("missing evaluation was updated")
	}
	if err := f.AddEvaluation(first); err != nil {
		t.Fatal(err)
	}
	newScore, err := NewScore(5)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := NewEvaluation(first.criterion, newScore)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.UpdateEvaluation(*updated); err != nil {
		t.Fatal(err)
	}
	if got := f.Evaluations(); len(got) != 1 || got[0].score.value != 5 {
		t.Fatal("update must replace the existing evaluation")
	}
	copy := f.Evaluations()
	copy[0] = second
	if f.Evaluations()[0].criterion.name != first.criterion.name {
		t.Fatal("returned slice exposes form state")
	}
	if err := f.RemoveEvaluation("missing"); err == nil {
		t.Fatal("missing evaluation was removed")
	}
}
