package memory_test

import (
	"context"
	"errors"
	"testing"

	"pos/internal/application"
	"pos/internal/domain"
	"pos/internal/infrastructure/memory"

	"github.com/google/uuid"
)

func rubricFixture(t *testing.T) (domain.Rubric, domain.Evaluation) {
	t.Helper()
	max, err := domain.NewScore(10)
	if err != nil {
		t.Fatal(err)
	}
	criterion, err := domain.NewCriterion("exam", max, 100)
	if err != nil {
		t.Fatal(err)
	}
	rubric, err := domain.NewRubric("Course", []domain.Criterion{criterion})
	if err != nil {
		t.Fatal(err)
	}
	evaluation, err := domain.NewEvaluation(criterion, max)
	if err != nil {
		t.Fatal(err)
	}
	evaluation.Comment = "original"
	return rubric, *evaluation
}

func TestStudentAndRubricRepositories(t *testing.T) {
	ctx := context.Background()
	students := memory.NewStudentRepository()
	student, err := domain.NewStudent("Alice")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := students.Get(ctx, student.ID()); !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("missing student: %v", err)
	}
	if err := students.Add(ctx, student); err != nil {
		t.Fatal(err)
	}
	if err := students.Add(ctx, student); !errors.Is(err, application.ErrAlreadyExists) {
		t.Fatalf("duplicate student: %v", err)
	}
	got, err := students.Get(ctx, student.ID())
	if err != nil || got != student {
		t.Fatalf("student changed: %v", err)
	}
	listed, err := students.List(ctx)
	if err != nil || len(listed) != 1 {
		t.Fatalf("list students: %v", err)
	}
	listed[0] = domain.Student{}
	if got, err := students.Get(ctx, student.ID()); err != nil || got.ID() != student.ID() {
		t.Fatal("list exposes student storage")
	}

	rubric, _ := rubricFixture(t)
	rubrics := memory.NewRubricRepository()
	if _, err := rubrics.Get(ctx, rubric.ID()); !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("missing rubric: %v", err)
	}
	if err := rubrics.Add(ctx, rubric); err != nil {
		t.Fatal(err)
	}
	if err := rubrics.Add(ctx, rubric); !errors.Is(err, application.ErrAlreadyExists) {
		t.Fatalf("duplicate rubric: %v", err)
	}
	stored, err := rubrics.Get(ctx, rubric.ID())
	if err != nil || stored.ID() != rubric.ID() {
		t.Fatalf("rubric ID changed: %v", err)
	}
	criteria := stored.Criteria()
	criteria[0] = domain.Criterion{}
	all, err := rubrics.List(ctx)
	if err != nil || len(all) != 1 || all[0].Criteria()[0].Name() != "exam" {
		t.Fatalf("rubric storage changed: %v", err)
	}
	all[0] = domain.Rubric{}
	if got, err := rubrics.Get(ctx, rubric.ID()); err != nil || got.Name() != rubric.Name() {
		t.Fatal("list exposes rubric storage")
	}
}

func TestFormRepositorySnapshots(t *testing.T) {
	ctx := context.Background()
	rubric, evaluation := rubricFixture(t)
	form, err := domain.NewEvaluationForm(uuid.New(), rubric)
	if err != nil {
		t.Fatal(err)
	}
	repo := memory.NewEvaluationFormRepository()
	if err := repo.Save(ctx, form); !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("save missing form: %v", err)
	}
	if err := repo.Create(ctx, form); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, form); !errors.Is(err, application.ErrAlreadyExists) {
		t.Fatalf("duplicate form: %v", err)
	}
	if err := form.AddEvaluation(evaluation); err != nil {
		t.Fatal(err)
	}
	stored, err := repo.Get(ctx, form.StudentID(), rubric.ID())
	if err != nil || len(stored.Evaluations()) != 0 {
		t.Fatal("input mutation changed stored form")
	}
	if err := form.Complete(); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(ctx, form); err != nil {
		t.Fatal(err)
	}
	if err := form.Reopen(); err != nil {
		t.Fatal(err)
	}
	if err := form.RemoveEvaluation("exam"); err != nil {
		t.Fatal(err)
	}
	stored, err = repo.Get(ctx, form.StudentID(), rubric.ID())
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status() != domain.Completed || len(stored.Evaluations()) != 1 {
		t.Fatal("save retained input alias")
	}
	gotEvaluation := stored.Evaluations()[0]
	if gotEvaluation.Comment != evaluation.Comment || !gotEvaluation.CreatedAt().Equal(evaluation.CreatedAt()) {
		t.Fatal("snapshot lost comment or creation time")
	}
	if err := stored.Reopen(); err != nil {
		t.Fatal(err)
	}
	if err := stored.RemoveEvaluation("exam"); err != nil {
		t.Fatal(err)
	}
	again, err := repo.Get(ctx, form.StudentID(), rubric.ID())
	if err != nil || again.Status() != domain.Completed || len(again.Evaluations()) != 1 {
		t.Fatal("read exposes stored form")
	}
	if _, err := repo.Get(ctx, form.StudentID(), uuid.New()); !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("wrong rubric: %v", err)
	}
	if err := repo.Create(ctx, nil); err == nil {
		t.Fatal("nil form accepted")
	}
}

func TestCanceledRepositoryOperations(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	rubric, _ := rubricFixture(t)
	form, err := domain.NewEvaluationForm(uuid.New(), rubric)
	if err != nil {
		t.Fatal(err)
	}
	students := memory.NewStudentRepository()
	rubrics := memory.NewRubricRepository()
	forms := memory.NewEvaluationFormRepository()
	operations := []func() error{
		func() error { return students.Add(ctx, domain.Student{}) },
		func() error { _, err := students.Get(ctx, uuid.New()); return err },
		func() error { _, err := students.List(ctx); return err },
		func() error { return rubrics.Add(ctx, rubric) },
		func() error { _, err := rubrics.Get(ctx, rubric.ID()); return err },
		func() error { _, err := rubrics.List(ctx); return err },
		func() error { return forms.Create(ctx, form) },
		func() error { _, err := forms.Get(ctx, form.StudentID(), rubric.ID()); return err },
		func() error { return forms.Save(ctx, form) },
	}
	for _, operation := range operations {
		if err := operation(); !errors.Is(err, context.Canceled) {
			t.Fatalf("lost cancellation: %v", err)
		}
	}
}
