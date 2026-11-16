package application_test

import (
	"testing"

	"pos/internal/application"
)

func TestRubricBuilderPreservesAcceptedSteps(t *testing.T) {
	builder, err := application.NewRubricBuilder(" Course ")
	if err != nil {
		t.Fatal(err)
	}
	if err := builder.AddCriterion(application.CriterionInput{Name: "Exam", MaxScore: 10, Weight: 40}); err != nil {
		t.Fatal(err)
	}
	if _, err := builder.Build(); err == nil {
		t.Fatal("incomplete rubric built")
	}
	for _, invalid := range []application.CriterionInput{
		{Name: "Exam", MaxScore: 10, Weight: 60},
		{Name: "Project", MaxScore: 0, Weight: 60},
		{Name: "Project", MaxScore: 10, Weight: 70},
	} {
		if err := builder.AddCriterion(invalid); err == nil {
			t.Fatalf("invalid step accepted: %+v", invalid)
		}
	}
	if err := builder.AddCriterion(application.CriterionInput{Name: "Project", MaxScore: 20, Weight: 60}); err != nil {
		t.Fatal(err)
	}
	rubric, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}
	if rubric.Name() != "Course" || len(rubric.Criteria()) != 2 || rubric.Criteria()[0].Name() != "Exam" {
		t.Fatal("builder lost accepted criteria")
	}
	if _, err := application.NewRubricBuilder(" \t"); err == nil {
		t.Fatal("blank builder name accepted")
	}
}
