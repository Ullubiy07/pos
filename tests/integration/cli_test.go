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

	"pos/internal/application"
	"pos/internal/cli"
	"pos/internal/di"
)

func cliFixture(input string, output *bytes.Buffer) (*cli.App, *application.RubricService) {
	container := di.New()
	return container.CLI(strings.NewReader(input), output), container.RubricService()
}

func TestCLIWorkflow(t *testing.T) {
	firstPath := filepath.Join(t.TempDir(), "first.csv")
	secondPath := filepath.Join(t.TempDir(), "second.csv")
	script := strings.Join([]string{
		"1", "Иван Петров",
		"3", "Course A", "2", "Exam", "100", "60", "Project", "10", "40",
		"5", "6", "7",
		"8", "1", "80", "Good", "8", "2", "5", "",
		"9", "11", "12", firstPath,
		"10", "8", "1", "1", "100", "Corrected", "8", "2", "3",
		"8", "2", "10", "", "9", "11",
		"3", "Course B", "1", "Exam", "20", "100",
		"6", "8", "1", "10", "", "9", "11", "12", secondPath,
		"4", "1", "7", "11", "2", "1", "0",
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
		"11", "unknown", "2", "4", // selection required, unknown command, empty lists
		"1", "Alice", "2", "bad", "2", "99", "2", "1",
		"3", "Invalid", "zero", // invalid integer
		"3", "Course", "1", "Exam", "10", "100",
		"6", "8", "1", "11", "", // score above maximum
		"8", "1", "7,5", "", "9", "11", "0",
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
		"6", "8", "1", "10", "", "9", "12", path, "0",
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

func TestCLIEditKeepsExistingValues(t *testing.T) {
	input := strings.Join([]string{
		"1", "Alice", "3", "Course", "1", "Exam", "10", "100",
		"6", "8", "1", "7", "Original",
		"8", "1", "2", "7",
		"8", "1", "1", "", "", "7",
		"8", "1", "1", "9", "", "7",
		"8", "1", "1", "", "New comment", "7",
		"8", "1", "1", "", "-", "7", "9", "11", "0",
	}, "\n") + "\n"
	var output bytes.Buffer
	app, _ := cliFixture(input, &output)
	if err := app.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	text := output.String()
	if strings.Contains(text, "Ошибка:") {
		t.Fatalf("unexpected error:\n%s", text)
	}
	for _, want := range []string{
		"Оценка оставлена без изменений.",
		"Балл [7]", "Балл [9]", "Комментарий [Original]", "Комментарий [New comment]",
		"Exam: 7 / 10; вес 100%; вклад 70; комментарий: Original",
		"Exam: 9 / 10; вес 100%; вклад 90; комментарий: Original",
		"Exam: 9 / 10; вес 100%; вклад 90; комментарий: New comment",
		"Exam: 9 / 10; вес 100%; вклад 90; комментарий: \n",
		"Итог: 90 / 100",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("missing output %q", want)
		}
	}
}

func TestCLIRubricBuilderRetriesOnlyRejectedCriterion(t *testing.T) {
	input := strings.Join([]string{
		"3", "Course", "2", "Exam", "10", "40",
		"Project", "20", "70", // exceeds the remaining weight; retry this step
		"Project", "20", "60", "0",
	}, "\n") + "\n"
	var output bytes.Buffer
	app, rubrics := cliFixture(input, &output)
	if err := app.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	listed, err := rubrics.List(context.Background())
	if err != nil || len(listed) != 1 || len(listed[0].Criteria()) != 2 {
		t.Fatalf("accepted criteria lost after retry: %v", err)
	}
	if !strings.Contains(output.String(), "Повторите ввод этого критерия") {
		t.Fatal("retry not explained")
	}
}

func TestCLIRoundingMatchesExport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rounded.csv")
	input := strings.Join([]string{
		"1", "Alice", "3", "Course", "1", "Exam", "3", "100",
		"6", "8", "1", "1", "", "9", "11", "12", path, "0",
	}, "\n") + "\n"
	var output bytes.Buffer
	app, _ := cliFixture(input, &output)
	if err := app.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Итог: 33.33 / 100") {
		t.Fatal("rounding policy not applied to CLI")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	records, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil || len(records) != 2 || records[1][4] != "33.33" {
		t.Fatalf("CSV uses a different rounding rule: %v", err)
	}
}
