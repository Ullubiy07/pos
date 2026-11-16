package csvexport

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"pos/internal/application"
)

type ResultExporter struct{}

var _ application.ResultExporter = (*ResultExporter)(nil)

func NewResultExporter() *ResultExporter { return &ResultExporter{} }

// Записывает заголовок и одну строку. Открытие и закрытие файла — на стороне вызывающего кода.
func (e *ResultExporter) Export(ctx context.Context, result application.StudentResult, destination io.Writer) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if destination == nil {
		return fmt.Errorf("CSV destination must not be nil")
	}
	writer := csv.NewWriter(destination)
	if err := writer.WriteAll([][]string{
		{"student_id", "student_name", "rubric_id", "rubric_name", "final_score"},
		{result.StudentID.String(), result.StudentName, result.RubricID.String(), result.RubricName,
			strconv.FormatFloat(result.FinalScore, 'f', -1, 64)},
	}); err != nil {
		return fmt.Errorf("write CSV result: %w", err)
	}
	return nil
}
