package domain

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
)

type EvaluationFormStatus string

const (
	Draft     EvaluationFormStatus = "Draft"
	Completed EvaluationFormStatus = "Completed"
)

type EvaluationForm struct {
	studentID   uuid.UUID
	rubric      Rubric
	status      EvaluationFormStatus
	evaluations []Evaluation
}

func NewEvaluationForm(studentID uuid.UUID, rubric Rubric) (*EvaluationForm, error) {
	if studentID == uuid.Nil {
		return nil, fmt.Errorf("student ID must not be empty")
	}
	validatedRubric, err := NewRubric(rubric.items)
	if err != nil {
		return nil, fmt.Errorf("invalid form rubric: %w", err)
	}
	return &EvaluationForm{
		studentID: studentID,
		rubric:    validatedRubric,
		status:    Draft,
	}, nil
}

func (f *EvaluationForm) Rubric() Rubric {
	return Rubric{items: f.rubric.Criteria()}
}

func (f *EvaluationForm) StudentID() uuid.UUID {
	return f.studentID
}

func (f *EvaluationForm) Status() EvaluationFormStatus {
	return f.status
}

func (f *EvaluationForm) Evaluations() []Evaluation {
	return slices.Clone(f.evaluations)
}

func (f *EvaluationForm) AddEvaluation(evaluation Evaluation) error {
	if err := f.validateEvaluation(evaluation); err != nil {
		return err
	}
	if f.evaluationIndex(evaluation.criterion.name) != -1 {
		return fmt.Errorf("evaluation for criterion %q already exists", evaluation.criterion.name)
	}
	f.evaluations = append(f.evaluations, evaluation)
	return nil
}

func (f *EvaluationForm) UpdateEvaluation(evaluation Evaluation) error {
	if err := f.validateEvaluation(evaluation); err != nil {
		return err
	}
	index := f.evaluationIndex(evaluation.criterion.name)
	if index == -1 {
		return fmt.Errorf("evaluation for criterion %q not found", evaluation.criterion.name)
	}
	f.evaluations[index] = evaluation
	return nil
}

func (f *EvaluationForm) RemoveEvaluation(criterionName string) error {
	if f.status != Draft {
		return fmt.Errorf("evaluations can only be changed in draft status")
	}
	index := f.evaluationIndex(criterionName)
	if index == -1 {
		return fmt.Errorf("evaluation for criterion %q not found", criterionName)
	}
	f.evaluations = slices.Delete(f.evaluations, index, index+1)
	return nil
}

func (f *EvaluationForm) Complete() error {
	if f.status != Draft {
		return fmt.Errorf("only a draft form can be completed")
	}
	if len(f.rubric.items) == 0 || len(f.evaluations) != len(f.rubric.items) {
		return fmt.Errorf("all rubric criteria must be evaluated before completion")
	}
	f.status = Completed
	return nil
}

func (f *EvaluationForm) Reopen() error {
	if f.status != Completed {
		return fmt.Errorf("only a completed form can be reopened")
	}
	f.status = Draft
	return nil
}

func (f *EvaluationForm) validateEvaluation(evaluation Evaluation) error {
	if f.status != Draft {
		return fmt.Errorf("evaluations can only be changed in draft status")
	}
	if !f.rubric.Contains(evaluation.criterion) {
		return fmt.Errorf("rubric does not contain evaluation criterion")
	}
	if _, err := NewEvaluation(evaluation.criterion, evaluation.score); err != nil {
		return fmt.Errorf("invalid evaluation: %w", err)
	}
	return nil
}

func (f *EvaluationForm) evaluationIndex(criterionName string) int {
	return slices.IndexFunc(f.evaluations, func(evaluation Evaluation) bool {
		return evaluation.criterion.name == criterionName
	})
}
