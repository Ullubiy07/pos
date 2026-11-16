package memory

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"

	"pos/internal/application"
	"pos/internal/domain"

	"github.com/google/uuid"
)

type StudentRepository struct {
	mu    sync.RWMutex
	items map[uuid.UUID]domain.Student
}

var _ application.StudentRepository = (*StudentRepository)(nil)

func NewStudentRepository() *StudentRepository {
	return &StudentRepository{items: make(map[uuid.UUID]domain.Student)}
}

func (r *StudentRepository) Add(ctx context.Context, student domain.Student) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if student.ID() == uuid.Nil {
		return fmt.Errorf("student ID must not be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[student.ID()]; exists {
		return fmt.Errorf("student %s: %w", student.ID(), application.ErrAlreadyExists)
	}
	if r.items == nil {
		r.items = make(map[uuid.UUID]domain.Student)
	}
	r.items[student.ID()] = student
	return nil
}

func (r *StudentRepository) Get(ctx context.Context, id uuid.UUID) (domain.Student, error) {
	if err := ctx.Err(); err != nil {
		return domain.Student{}, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	student, exists := r.items[id]
	if !exists {
		return domain.Student{}, fmt.Errorf("student %s: %w", id, application.ErrNotFound)
	}
	return student, nil
}

func (r *StudentRepository) List(ctx context.Context) ([]domain.Student, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	students := make([]domain.Student, 0, len(r.items))
	for _, student := range r.items {
		students = append(students, student)
	}
	slices.SortFunc(students, func(a, b domain.Student) int {
		if byName := strings.Compare(a.Name(), b.Name()); byName != 0 {
			return byName
		}
		return strings.Compare(a.ID().String(), b.ID().String())
	})
	return students, nil
}
