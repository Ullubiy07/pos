package main

import (
	"context"
	"fmt"
	"os"

	"pos/internal/cli"
	"pos/internal/application"
	"pos/internal/domain"
	"pos/internal/infrastructure/csvexport"
	"pos/internal/infrastructure/memory"
)

func main() {
	students := memory.NewStudentRepository()
	rubrics := memory.NewRubricRepository()
	forms := memory.NewEvaluationFormRepository()
	app := cli.New(os.Stdin, os.Stdout,
		application.NewStudentService(students),
		application.NewRubricService(rubrics),
		application.NewEvaluationFormService(students, rubrics, forms),
		application.NewResultService(students, forms, domain.NewEvaluationService(), csvexport.NewResultExporter()),
	)
	if err := app.Run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
