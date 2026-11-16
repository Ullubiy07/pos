package application

import (
	"context"
	"io"

	"pos/internal/domain"

	"github.com/google/uuid"
)

// Facade is the CLI's entry point to application scenarios. Services retain
// orchestration and domain rules remain in Domain.
type Facade struct {
	students *StudentService
	rubrics  *RubricService
	forms    *EvaluationFormService
	results  *ResultService
}

func NewFacade(students *StudentService, rubrics *RubricService, forms *EvaluationFormService, results *ResultService) *Facade {
	return &Facade{students: students, rubrics: rubrics, forms: forms, results: results}
}

func (f *Facade) AddStudent(ctx context.Context, name string) (domain.Student, error) {
	return f.students.Add(ctx, name)
}

func (f *Facade) ListStudents(ctx context.Context) ([]domain.Student, error) {
	return f.students.List(ctx)
}

func (f *Facade) SaveRubric(ctx context.Context, builder *RubricBuilder) (domain.Rubric, error) {
	return f.rubrics.Save(ctx, builder)
}

func (f *Facade) ListRubrics(ctx context.Context) ([]domain.Rubric, error) {
	return f.rubrics.List(ctx)
}

func (f *Facade) GetRubric(ctx context.Context, id uuid.UUID) (domain.Rubric, error) {
	return f.rubrics.Get(ctx, id)
}

func (f *Facade) CreateForm(ctx context.Context, studentID, rubricID uuid.UUID) (*domain.EvaluationForm, error) {
	return f.forms.Create(ctx, studentID, rubricID)
}

func (f *Facade) GetForm(ctx context.Context, studentID, rubricID uuid.UUID) (*domain.EvaluationForm, error) {
	return f.forms.Get(ctx, studentID, rubricID)
}

func (f *Facade) AddEvaluation(ctx context.Context, studentID, rubricID uuid.UUID, criterion string, score float64, comment string) error {
	return f.forms.AddEvaluation(ctx, studentID, rubricID, criterion, score, comment)
}

func (f *Facade) UpdateEvaluation(ctx context.Context, studentID, rubricID uuid.UUID, criterion string, score float64, comment string) error {
	return f.forms.UpdateEvaluation(ctx, studentID, rubricID, criterion, score, comment)
}

func (f *Facade) RemoveEvaluation(ctx context.Context, studentID, rubricID uuid.UUID, criterion string) error {
	return f.forms.RemoveEvaluation(ctx, studentID, rubricID, criterion)
}

func (f *Facade) CompleteForm(ctx context.Context, studentID, rubricID uuid.UUID) error {
	return f.forms.Complete(ctx, studentID, rubricID)
}

func (f *Facade) ReopenForm(ctx context.Context, studentID, rubricID uuid.UUID) error {
	return f.forms.Reopen(ctx, studentID, rubricID)
}

func (f *Facade) CalculateResult(ctx context.Context, studentID, rubricID uuid.UUID) (domain.Score, error) {
	return f.results.Calculate(ctx, studentID, rubricID)
}

func (f *Facade) ExportResult(ctx context.Context, studentID, rubricID uuid.UUID, destination io.Writer) error {
	return f.results.Export(ctx, studentID, rubricID, destination)
}
