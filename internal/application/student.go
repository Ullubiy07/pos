package application

import (
	"context"
	"fmt"

	"pos/internal/domain"

	"github.com/google/uuid"
)

type StudentService struct {
	students StudentRepository
}

func NewStudentService(students StudentRepository) *StudentService {
	return &StudentService{students: students}
}

func (s *StudentService) Add(ctx context.Context, name string) (domain.Student, error) {
	student, err := domain.NewStudent(name)
	if err != nil {
		return domain.Student{}, fmt.Errorf("create student: %w", err)
	}
	if err := s.students.Add(ctx, student); err != nil {
		return domain.Student{}, fmt.Errorf("save student: %w", err)
	}
	return student, nil
}

func (s *StudentService) Get(ctx context.Context, id uuid.UUID) (domain.Student, error) {
	student, err := s.students.Get(ctx, id)
	if err != nil {
		return domain.Student{}, fmt.Errorf("get student: %w", err)
	}
	return student, nil
}

func (s *StudentService) List(ctx context.Context) ([]domain.Student, error) {
	students, err := s.students.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list students: %w", err)
	}
	return students, nil
}
