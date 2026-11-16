package di_test

import (
	"context"
	"errors"
	"testing"

	"pos/internal/application"
	"pos/internal/di"
)

func TestContainersKeepSeparateStorage(t *testing.T) {
	ctx := context.Background()
	first, second := di.New(), di.New()
	student, err := first.StudentService().Add(ctx, "Alice")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := first.StudentService().Get(ctx, student.ID()); err != nil {
		t.Fatalf("repeated resolution lost stored student: %v", err)
	}
	if _, err := second.StudentService().Get(ctx, student.ID()); !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("containers share storage: %v", err)
	}
	rubric, err := first.RubricService().Create(ctx, "Course", []application.CriterionInput{
		{Name: "Exam", MaxScore: 10, Weight: 100},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := first.EvaluationFormService().Create(ctx, student.ID(), rubric.ID()); err != nil {
		t.Fatalf("services do not share the container's repositories: %v", err)
	}
}
