package application_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"pos/internal/application"
	"pos/internal/domain"

	"github.com/google/uuid"
)

type studentStore struct {
	items map[uuid.UUID]domain.Student
	err   error
}

func (s *studentStore) Add(_ context.Context, student domain.Student) error {
	if s.err != nil {
		return s.err
	}
	if _, exists := s.items[student.ID()]; exists {
		return application.ErrAlreadyExists
	}
	s.items[student.ID()] = student
	return nil
}
func (s *studentStore) Get(_ context.Context, id uuid.UUID) (domain.Student, error) {
	if s.err != nil {
		return domain.Student{}, s.err
	}
	student, ok := s.items[id]
	if !ok {
		return domain.Student{}, application.ErrNotFound
	}
	return student, nil
}
func (s *studentStore) List(_ context.Context) ([]domain.Student, error) {
	if s.err != nil {
		return nil, s.err
	}
	var students []domain.Student
	for _, student := range s.items {
		students = append(students, student)
	}
	return students, nil
}

type rubricStore struct {
	items map[uuid.UUID]domain.Rubric
	err   error
}

func (s *rubricStore) Add(_ context.Context, rubric domain.Rubric) error {
	if s.err != nil {
		return s.err
	}
	if _, exists := s.items[rubric.ID()]; exists {
		return application.ErrAlreadyExists
	}
	s.items[rubric.ID()] = rubric
	return nil
}
func (s *rubricStore) Get(_ context.Context, id uuid.UUID) (domain.Rubric, error) {
	if s.err != nil {
		return domain.Rubric{}, s.err
	}
	rubric, ok := s.items[id]
	if !ok {
		return domain.Rubric{}, application.ErrNotFound
	}
	return rubric, nil
}
func (s *rubricStore) List(context.Context) ([]domain.Rubric, error) {
	if s.err != nil {
		return nil, s.err
	}
	var rubrics []domain.Rubric
	for _, rubric := range s.items {
		rubrics = append(rubrics, rubric)
	}
	return rubrics, nil
}

type formKey struct {
	studentID uuid.UUID
	rubricID  uuid.UUID
}

type formStore struct {
	items   map[formKey]*domain.EvaluationForm
	saveErr error
	saves   int
}

func cloneForm(form *domain.EvaluationForm) *domain.EvaluationForm {
	copy, err := domain.NewEvaluationForm(form.StudentID(), form.Rubric())
	if err != nil {
		panic(err)
	}
	for _, evaluation := range form.Evaluations() {
		if err := copy.AddEvaluation(evaluation); err != nil {
			panic(err)
		}
	}
	if form.Status() == domain.Completed {
		if err := copy.Complete(); err != nil {
			panic(err)
		}
	}
	return copy
}
func (s *formStore) Create(_ context.Context, form *domain.EvaluationForm) error {
	if _, exists := s.items[formKey{form.StudentID(), form.RubricID()}]; exists {
		return application.ErrAlreadyExists
	}
	s.items[formKey{form.StudentID(), form.RubricID()}] = cloneForm(form)
	return nil
}
func (s *formStore) Get(_ context.Context, studentID, rubricID uuid.UUID) (*domain.EvaluationForm, error) {
	form, ok := s.items[formKey{studentID, rubricID}]
	if !ok {
		return nil, application.ErrNotFound
	}
	return cloneForm(form), nil
}
func (s *formStore) Save(_ context.Context, form *domain.EvaluationForm) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	if _, exists := s.items[formKey{form.StudentID(), form.RubricID()}]; !exists {
		return application.ErrNotFound
	}
	s.items[formKey{form.StudentID(), form.RubricID()}] = cloneForm(form)
	s.saves++
	return nil
}

type resultExporter struct {
	result application.StudentResult
	err    error
	calls  int
}

func (e *resultExporter) Export(_ context.Context, result application.StudentResult, writer io.Writer) error {
	e.calls++
	e.result = result
	if e.err != nil {
		return e.err
	}
	_, err := io.WriteString(writer, "exported")
	return err
}

func TestApplicationEvaluationWorkflow(t *testing.T) {
	ctx := context.Background()
	students := &studentStore{items: make(map[uuid.UUID]domain.Student)}
	rubrics := &rubricStore{items: make(map[uuid.UUID]domain.Rubric)}
	forms := &formStore{items: make(map[formKey]*domain.EvaluationForm)}
	exporter := &resultExporter{}
	studentService := application.NewStudentService(students)
	rubricService := application.NewRubricService(rubrics)
	formService := application.NewEvaluationFormService(students, rubrics, forms)
	results := application.NewResultService(students, forms, domain.NewEvaluationService(domain.ExactScorePolicy{}), exporter)

	student, err := studentService.Add(ctx, " Alice ")
	if err != nil {
		t.Fatal(err)
	}
	if listed, err := studentService.List(ctx); err != nil || len(listed) != 1 {
		t.Fatalf("list students: %v, %v", listed, err)
	}

	if _, err := rubricService.Create(ctx, "Incomplete", []application.CriterionInput{{Name: "exam", MaxScore: 100, Weight: 60}}); err == nil {
		t.Fatal("incomplete rubric accepted")
	}
	rubric, err := rubricService.Create(ctx, "Course", []application.CriterionInput{
		{Name: "exam", MaxScore: 100, Weight: 60},
		{Name: "project", MaxScore: 10, Weight: 40},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rubricService.Create(ctx, "Duplicates", []application.CriterionInput{
		{Name: "exam", MaxScore: 100, Weight: 60},
		{Name: " exam ", MaxScore: 10, Weight: 40},
	}); err == nil {
		t.Fatal("duplicate criterion accepted")
	}
	if got, err := rubricService.Get(ctx, rubric.ID()); err != nil || got.Name() != "Course" {
		t.Fatalf("get rubric: %v", err)
	}
	if _, err := formService.Create(ctx, uuid.New(), rubric.ID()); !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("unknown student: %v", err)
	}
	if _, err := formService.Create(ctx, student.ID(), uuid.New()); !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("unknown rubric: %v", err)
	}
	if _, err := formService.Create(ctx, student.ID(), rubric.ID()); err != nil {
		t.Fatal(err)
	}
	if _, err := formService.Create(ctx, student.ID(), rubric.ID()); !errors.Is(err, application.ErrAlreadyExists) {
		t.Fatalf("duplicate form: %v", err)
	}

	for _, tc := range []struct {
		name  string
		score float64
	}{{"missing", 5}, {"exam", 101}, {"exam", -1}} {
		if err := formService.AddEvaluation(ctx, student.ID(), rubric.ID(), tc.name, tc.score, ""); err == nil {
			t.Fatalf("invalid evaluation accepted: %+v", tc)
		}
	}
	if forms.saves != 0 {
		t.Fatal("invalid evaluations were saved")
	}
	if err := formService.AddEvaluation(ctx, student.ID(), rubric.ID(), " exam ", 80, "good"); err != nil {
		t.Fatal(err)
	}
	if err := formService.AddEvaluation(ctx, student.ID(), rubric.ID(), "exam", 80, ""); err == nil {
		t.Fatal("duplicate evaluation accepted")
	}
	if err := formService.Complete(ctx, student.ID(), rubric.ID()); err == nil {
		t.Fatal("incomplete form completed")
	}
	if err := formService.AddEvaluation(ctx, student.ID(), rubric.ID(), "project", 5, ""); err != nil {
		t.Fatal(err)
	}
	if err := formService.Complete(ctx, student.ID(), rubric.ID()); err != nil {
		t.Fatal(err)
	}
	score, err := results.Calculate(ctx, student.ID(), rubric.ID())
	if err != nil || score.Value() != 68 {
		t.Fatalf("score: %v, %v; want 68", score.Value(), err)
	}
	if err := formService.UpdateEvaluation(ctx, student.ID(), rubric.ID(), "exam", 100, ""); err == nil {
		t.Fatal("completed form changed")
	}

	var output bytes.Buffer
	if err := results.Export(ctx, student.ID(), rubric.ID(), &output); err != nil {
		t.Fatal(err)
	}
	if exporter.result.StudentID != student.ID() || exporter.result.StudentName != "Alice" || exporter.result.FinalScore != 68 || output.String() != "exported" {
		t.Fatalf("unexpected export: %+v, %q", exporter.result, output.String())
	}

	if err := formService.Reopen(ctx, student.ID(), rubric.ID()); err != nil {
		t.Fatal(err)
	}
	if err := results.Export(ctx, student.ID(), rubric.ID(), &output); err == nil || exporter.calls != 1 {
		t.Fatal("draft result was exported")
	}
	if err := formService.UpdateEvaluation(ctx, student.ID(), rubric.ID(), "exam", 100, "corrected"); err != nil {
		t.Fatal(err)
	}
	if err := formService.RemoveEvaluation(ctx, student.ID(), rubric.ID(), "project"); err != nil {
		t.Fatal(err)
	}
	if err := formService.AddEvaluation(ctx, student.ID(), rubric.ID(), "project", 10, ""); err != nil {
		t.Fatal(err)
	}
	if err := formService.Complete(ctx, student.ID(), rubric.ID()); err != nil {
		t.Fatal(err)
	}
	score, err = results.Calculate(ctx, student.ID(), rubric.ID())
	if err != nil || score.Value() != 100 {
		t.Fatalf("corrected score: %v, %v", score.Value(), err)
	}
}

func TestApplicationPreservesInfrastructureErrors(t *testing.T) {
	ctx := context.Background()
	storageErr := errors.New("storage unavailable")
	students := &studentStore{items: make(map[uuid.UUID]domain.Student), err: storageErr}
	studentService := application.NewStudentService(students)
	if _, err := studentService.Add(ctx, "Alice"); !errors.Is(err, storageErr) {
		t.Fatalf("lost storage error: %v", err)
	}
	students.err = nil
	student, err := studentService.Add(ctx, "Alice")
	if err != nil {
		t.Fatal(err)
	}
	rubrics := &rubricStore{items: make(map[uuid.UUID]domain.Rubric)}
	rubric, err := application.NewRubricService(rubrics).Create(ctx, "Course", []application.CriterionInput{{Name: "exam", MaxScore: 10, Weight: 100}})
	if err != nil {
		t.Fatal(err)
	}
	forms := &formStore{items: make(map[formKey]*domain.EvaluationForm), saveErr: storageErr}
	formService := application.NewEvaluationFormService(students, rubrics, forms)
	if _, err := formService.Create(ctx, student.ID(), rubric.ID()); err != nil {
		t.Fatal(err)
	}
	if err := formService.AddEvaluation(ctx, student.ID(), rubric.ID(), "exam", 10, ""); !errors.Is(err, storageErr) {
		t.Fatalf("lost save error: %v", err)
	}
	form, err := formService.Get(ctx, student.ID(), rubric.ID())
	if err != nil || len(form.Evaluations()) != 0 {
		t.Fatal("failed save changed stored form")
	}
	forms.saveErr = nil
	if err := formService.AddEvaluation(ctx, student.ID(), rubric.ID(), "exam", 10, ""); err != nil {
		t.Fatal(err)
	}
	if err := formService.Complete(ctx, student.ID(), rubric.ID()); err != nil {
		t.Fatal(err)
	}
	exporter := &resultExporter{err: storageErr}
	results := application.NewResultService(students, forms, domain.NewEvaluationService(domain.ExactScorePolicy{}), exporter)
	if err := results.Export(ctx, student.ID(), rubric.ID(), io.Discard); !errors.Is(err, storageErr) {
		t.Fatalf("lost export error: %v", err)
	}
}

func TestFormsUseSelectedRubric(t *testing.T) {
	ctx := context.Background()
	students := &studentStore{items: make(map[uuid.UUID]domain.Student)}
	rubrics := &rubricStore{items: make(map[uuid.UUID]domain.Rubric)}
	forms := &formStore{items: make(map[formKey]*domain.EvaluationForm)}
	studentService := application.NewStudentService(students)
	rubricService := application.NewRubricService(rubrics)
	formService := application.NewEvaluationFormService(students, rubrics, forms)

	first, err := rubricService.Create(ctx, "First course", []application.CriterionInput{
		{Name: "exam", MaxScore: 10, Weight: 100},
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := rubricService.Create(ctx, "Second course", []application.CriterionInput{
		{Name: "exam", MaxScore: 100, Weight: 100},
	})
	if err != nil {
		t.Fatal(err)
	}
	if first.ID() == second.ID() {
		t.Fatal("different rubrics have the same ID")
	}
	listed, err := rubricService.List(ctx)
	if err != nil || len(listed) != 2 {
		t.Fatalf("list rubrics: %v, %v", listed, err)
	}
	student, err := studentService.Add(ctx, "Student")
	if err != nil {
		t.Fatal(err)
	}
	for i, rubric := range []domain.Rubric{first, second} {
		form, err := formService.Create(ctx, student.ID(), rubric.ID())
		if err != nil {
			t.Fatal(err)
		}
		if form.Rubric().ID() != rubric.ID() || form.Rubric().Name() != rubric.Name() {
			t.Fatal("form lost selected rubric identity")
		}
		err = formService.AddEvaluation(ctx, student.ID(), rubric.ID(), "exam", 50, "")
		if (i == 0 && err == nil) || (i == 1 && err != nil) {
			t.Fatalf("form used wrong rubric maximum: %v", err)
		}
		stored, err := formService.Get(ctx, student.ID(), rubric.ID())
		if err != nil || stored.Rubric().ID() != rubric.ID() {
			t.Fatalf("stored rubric: %v", err)
		}
	}
	if _, err := formService.Create(ctx, student.ID(), first.ID()); !errors.Is(err, application.ErrAlreadyExists) {
		t.Fatalf("duplicate pair accepted: %v", err)
	}
	if _, err := formService.Get(ctx, student.ID(), uuid.New()); !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("missing pair: %v", err)
	}
	if err := formService.AddEvaluation(ctx, student.ID(), first.ID(), "exam", 10, ""); err != nil {
		t.Fatal(err)
	}
	for _, rubric := range []domain.Rubric{first, second} {
		if err := formService.Complete(ctx, student.ID(), rubric.ID()); err != nil {
			t.Fatal(err)
		}
	}
	exporter := &resultExporter{}
	results := application.NewResultService(students, forms, domain.NewEvaluationService(domain.ExactScorePolicy{}), exporter)
	for i, rubric := range []domain.Rubric{first, second} {
		want := []float64{100, 50}[i]
		score, err := results.Calculate(ctx, student.ID(), rubric.ID())
		if err != nil || score.Value() != want {
			t.Fatalf("rubric %q score: %v, %v; want %v", rubric.Name(), score.Value(), err, want)
		}
		if err := results.Export(ctx, student.ID(), rubric.ID(), io.Discard); err != nil {
			t.Fatal(err)
		}
		if exporter.result.RubricID != rubric.ID() || exporter.result.RubricName != rubric.Name() || exporter.result.FinalScore != want {
			t.Fatalf("wrong rubric exported: %+v", exporter.result)
		}
	}
	if err := formService.Reopen(ctx, student.ID(), first.ID()); err != nil {
		t.Fatal(err)
	}
	if err := formService.RemoveEvaluation(ctx, student.ID(), first.ID(), "exam"); err != nil {
		t.Fatal(err)
	}
	secondForm, err := formService.Get(ctx, student.ID(), second.ID())
	if err != nil || secondForm.Status() != domain.Completed || len(secondForm.Evaluations()) != 1 {
		t.Fatalf("changing first form affected second form: %v", err)
	}

	storageErr := errors.New("rubric storage unavailable")
	rubrics.err = storageErr
	if _, err := rubricService.Create(ctx, "Third", []application.CriterionInput{{Name: "exam", MaxScore: 10, Weight: 100}}); !errors.Is(err, storageErr) {
		t.Fatalf("lost rubric save error: %v", err)
	}
	if _, err := rubricService.Get(ctx, first.ID()); !errors.Is(err, storageErr) {
		t.Fatalf("lost rubric get error: %v", err)
	}
	if _, err := rubricService.List(ctx); !errors.Is(err, storageErr) {
		t.Fatalf("lost rubric list error: %v", err)
	}
}
