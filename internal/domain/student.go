package domain

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Student struct {
	id   uuid.UUID
	name string
}

func NewStudent(name string) (Student, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Student{}, fmt.Errorf("student name must be non-empty")
	}
	return Student{
		name: name,
		id:   uuid.New(),
	}, nil
}

func (s Student) ID() uuid.UUID { return s.id }

func (s Student) Name() string { return s.name }
