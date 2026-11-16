package integration_test

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pos/internal/cli"
	"pos/internal/application"
	"pos/internal/domain"
	"pos/internal/infrastructure/csvexport"
	"pos/internal/infrastructure/memory"
)

func cliFixture(input string, output *bytes.Buffer) (*cli.App, *application.RubricService) {
	students := memory.NewStudentRepository()
	rubrics := memory.NewRubricRepository()
	forms := memory.NewEvaluationFormRepository()
	rubricService := application.NewRubricService(rubrics)
	return cli.New(strings.NewReader(input), output,
		application.NewStudentService(students), rubricService,
		application.NewEvaluationFormService(students, rubrics, forms),
		application.NewResultService(students, forms, domain.NewEvaluationService(), csvexport.NewResultExporter()),
	), rubricService
}

func TestCLIWorkflow(t *testing.T) {
	firstPath := filepath.Join(t.TempDir(), "first.csv")
	secondPath := filepath.Join(t.TempDir(), "second.csv")
	script := strings.Join([]string{
		"1", "Иван Петров",
		"3", "Course A", "2", "Exam", "100", "60", "Project", "10", "40",
		"5", "6", "7",
		"8", "1", "80", "Good", "8", "2", "5", "",
		"11", "13", "14", firstPath,
		"12", "9", "1", "100", "Corrected", "10", "2",
		"8", "2", "10", "", "11", "13",
		"3", "Course B", "1", "Exam", "20", "100",
		"6", "8", "1", "10", "", "11", "13", "14", secondPath,
		"4", "1", "7", "13", "2", "1", "0",
	}, "\n") + "\n"
	var output bytes.Buffer
	app, _ := cliFixture(script, &output)
	if err := app.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	text := output.String()
	if strings.Contains(text, "Ошибка:") {
		t.Fatalf("unexpected CLI error:\n%s", text)
	}
	for _, want := range []string{"Итог: 68 / 100", "Итог: 100 / 100", "Итог: 50 / 100", "Corrected", "CSV сохранён:"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing output %q", want)
		}
	}
	for i, path := range []string{firstPath, secondPath} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		records, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
		if err != nil || len(records) != 2 {
			t.Fatalf("invalid exported CSV: %v", err)
		}
		if records[1][1] != "Иван Петров" || records[1][4] != []string{"68", "50"}[i] {
			t.Fatalf("wrong CSV row: %v", records[1])
		}
	}
}

func TestCLIRecoversFromInputErrors(t *testing.T) {
	input := strings.Join([]string{
		"13", "unknown", "2", "4", // selection required, unknown command, empty lists
		"1", "Alice", "2", "bad", "2", "99", "2", "1",
		"3", "Invalid", "zero", // invalid integer
		"3", "Course", "1", "Exam", "10", "100",
		"6", "8", "1", "11", "", // score above maximum
		"8", "1", "7,5", "", "11", "13", "0",
	}, "\n") + "\n"
	var output bytes.Buffer
	app, _ := cliFixture(input, &output)
	if err := app.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Ошибка:") || !strings.Contains(output.String(), "Итог: 75 / 100") {
		t.Fatalf("CLI did not recover:\n%s", output.String())
	}
}

func TestCLIEOFDoesNotSavePartialRubric(t *testing.T) {
	var output bytes.Buffer
	app, rubrics := cliFixture("3\nCourse\n2\nExam\n10\n50\n", &output)
	if err := app.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	listed, err := rubrics.List(context.Background())
	if err != nil || len(listed) != 0 {
		t.Fatalf("partial rubric saved: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	app, _ = cliFixture("", &output)
	if err := app.Run(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("lost cancellation: %v", err)
	}
}

func TestCLIExportDoesNotOverwriteExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "result.csv")
	if err := os.WriteFile(path, []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}
	input := strings.Join([]string{
		"1", "Alice", "3", "Course", "1", "Exam", "10", "100",
		"6", "8", "1", "10", "", "11", "14", path, "0",
	}, "\n") + "\n"
	var output bytes.Buffer
	app, _ := cliFixture(input, &output)
	if err := app.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "existing" {
		t.Fatal("existing file was modified")
	}
	if !strings.Contains(output.String(), "Ошибка:") {
		t.Fatal("overwrite error not shown")
	}
}
