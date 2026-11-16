package integration_test

import (
	"bytes"
	"context"
	"encoding/csv"
	"testing"

	"pos/internal/application"
	"pos/internal/domain"
	"pos/internal/infrastructure/csvexport"
	"pos/internal/infrastructure/memory"
)

func TestServicesWithMemoryRepositoriesAndCSV(t *testing.T) {
	ctx := context.Background()
	students := memory.NewStudentRepository()
	rubrics := memory.NewRubricRepository()
	forms := memory.NewEvaluationFormRepository()
	studentService := application.NewStudentService(students)
	rubricService := application.NewRubricService(rubrics)
	formService := application.NewEvaluationFormService(students, rubrics, forms)
	results := application.NewResultService(students, forms, domain.NewEvaluationService(domain.ExactScorePolicy{}), csvexport.NewResultExporter())
	student, err := studentService.Add(ctx, "Иван, Петров")
	if err != nil {
		t.Fatal(err)
	}
	for i, max := range []float64{10, 100} {
		rubric, err := rubricService.Create(ctx, "Course", []application.CriterionInput{{Name: "exam", MaxScore: max, Weight: 100}})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := formService.Create(ctx, student.ID(), rubric.ID()); err != nil {
			t.Fatal(err)
		}
		if err := formService.AddEvaluation(ctx, student.ID(), rubric.ID(), "exam", max/2, "half"); err != nil {
			t.Fatal(err)
		}
		if err := formService.Complete(ctx, student.ID(), rubric.ID()); err != nil {
			t.Fatal(err)
		}
		var output bytes.Buffer
		if err := results.Export(ctx, student.ID(), rubric.ID(), &output); err != nil {
			t.Fatal(err)
		}
		records, err := csv.NewReader(&output).ReadAll()
		if err != nil || len(records) != 2 {
			t.Fatalf("CSV %d: %v, %v", i, records, err)
		}
		if records[1][0] != student.ID().String() || records[1][1] != student.Name() || records[1][2] != rubric.ID().String() || records[1][4] != "50" {
			t.Fatalf("wrong exported result: %v", records[1])
		}
	}
}
