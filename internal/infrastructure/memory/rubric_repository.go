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

type RubricRepository struct {
	mu    sync.RWMutex
	items map[uuid.UUID]domain.Rubric
}

var _ application.RubricRepository = (*RubricRepository)(nil)

func NewRubricRepository() *RubricRepository {
	return &RubricRepository{items: make(map[uuid.UUID]domain.Rubric)}
}

func (r *RubricRepository) Add(ctx context.Context, rubric domain.Rubric) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if rubric.ID() == uuid.Nil {
		return fmt.Errorf("rubric ID must not be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[rubric.ID()]; exists {
		return fmt.Errorf("rubric %s: %w", rubric.ID(), application.ErrAlreadyExists)
	}
	if r.items == nil {
		r.items = make(map[uuid.UUID]domain.Rubric)
	}
	r.items[rubric.ID()] = rubric
	return nil
}

func (r *RubricRepository) Get(ctx context.Context, id uuid.UUID) (domain.Rubric, error) {
	if err := ctx.Err(); err != nil {
		return domain.Rubric{}, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	rubric, exists := r.items[id]
	if !exists {
		return domain.Rubric{}, fmt.Errorf("rubric %s: %w", id, application.ErrNotFound)
	}
	return rubric, nil
}

func (r *RubricRepository) List(ctx context.Context) ([]domain.Rubric, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	rubrics := make([]domain.Rubric, 0, len(r.items))
	for _, rubric := range r.items {
		rubrics = append(rubrics, rubric)
	}
	slices.SortFunc(rubrics, func(a, b domain.Rubric) int {
		if byName := strings.Compare(a.Name(), b.Name()); byName != 0 {
			return byName
		}
		return strings.Compare(a.ID().String(), b.ID().String())
	})
	return rubrics, nil
}
