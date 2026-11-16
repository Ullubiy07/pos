package csvexport_test

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"io"
	"reflect"
	"testing"

	"pos/internal/application"
	"pos/internal/infrastructure/csvexport"

	"github.com/google/uuid"
)

func TestResultExporterEscapesCSV(t *testing.T) {
	result := application.StudentResult{
		StudentID: uuid.New(), StudentName: "Иван, \"Петров\"\nСтудент",
		RubricID: uuid.New(), RubricName: "Курс, \"Go\"", FinalScore: 82.5,
	}
	var output bytes.Buffer
	if err := csvexport.NewResultExporter().Export(context.Background(), result, &output); err != nil {
		t.Fatal(err)
	}
	records, err := csv.NewReader(&output).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"student_id", "student_name", "rubric_id", "rubric_name", "final_score"},
		{result.StudentID.String(), result.StudentName, result.RubricID.String(), result.RubricName, "82.5"},
	}
	if !reflect.DeepEqual(records, want) {
		t.Fatalf("got %q, want %q", records, want)
	}
}

type failedWriter struct{ err error }

func (w failedWriter) Write([]byte) (int, error) { return 0, w.err }

func TestResultExporterErrors(t *testing.T) {
	exporter := csvexport.NewResultExporter()
	writeErr := errors.New("disk full")
	if err := exporter.Export(context.Background(), application.StudentResult{}, failedWriter{writeErr}); !errors.Is(err, writeErr) {
		t.Fatalf("lost buffered write error: %v", err)
	}
	if err := exporter.Export(context.Background(), application.StudentResult{}, nil); err == nil {
		t.Fatal("nil destination accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := exporter.Export(ctx, application.StudentResult{}, io.Discard); !errors.Is(err, context.Canceled) {
		t.Fatalf("lost cancellation: %v", err)
	}
}
