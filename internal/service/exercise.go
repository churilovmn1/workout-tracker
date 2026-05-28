package service

import (
	"context"

	"github.com/churilovmn1/workout-tracker/internal/models"
)

type exerciseRepository interface {
	Create(ctx context.Context, ex *models.Exercise) (int, error)
	GetByID(ctx context.Context, id int) (*models.Exercise, error)
	List(ctx context.Context, muscleGroup, search string) ([]models.Exercise, error)
	Update(ctx context.Context, ex *models.Exercise) error
	Delete(ctx context.Context, id int) error
}

// ExerciseService handles exercise catalog business logic.
type ExerciseService struct {
	repo exerciseRepository
}

// NewExerciseService creates a new ExerciseService.
func NewExerciseService(repo exerciseRepository) *ExerciseService {
	return &ExerciseService{repo: repo}
}

// Create adds a new exercise to the catalog.
func (s *ExerciseService) Create(ctx context.Context, ex *models.Exercise) (int, error) {
	return s.repo.Create(ctx, ex)
}

// GetByID returns an exercise by its ID.
func (s *ExerciseService) GetByID(ctx context.Context, id int) (*models.Exercise, error) {
	return s.repo.GetByID(ctx, id)
}

// List returns exercises filtered by muscle group and/or name search term.
func (s *ExerciseService) List(ctx context.Context, muscleGroup, search string) ([]models.Exercise, error) {
	return s.repo.List(ctx, muscleGroup, search)
}

// Update modifies an existing exercise.
func (s *ExerciseService) Update(ctx context.Context, ex *models.Exercise) error {
	return s.repo.Update(ctx, ex)
}

// Delete removes an exercise by ID.
func (s *ExerciseService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
