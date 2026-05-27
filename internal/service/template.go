package service

import (
	"context"

	"github.com/churilovmn1/workout-tracker/internal/models"
	"github.com/churilovmn1/workout-tracker/internal/repository"
)

// TemplateService handles workout template business logic.
type TemplateService struct {
	repo *repository.TemplateRepository
}

// NewTemplateService creates a new TemplateService.
func NewTemplateService(repo *repository.TemplateRepository) *TemplateService {
	return &TemplateService{repo: repo}
}

// Create adds a new workout template.
func (s *TemplateService) Create(ctx context.Context, t *models.WorkoutTemplate) (int, error) {
	return s.repo.Create(ctx, t)
}

// GetByID returns a template by ID.
func (s *TemplateService) GetByID(ctx context.Context, id int) (*models.WorkoutTemplate, error) {
	return s.repo.GetByID(ctx, id)
}

// ListByUser returns templates available to the user.
func (s *TemplateService) ListByUser(ctx context.Context, userID int) ([]models.WorkoutTemplate, error) {
	return s.repo.ListByUser(ctx, userID)
}

// Update modifies a template owned by the user.
func (s *TemplateService) Update(ctx context.Context, t *models.WorkoutTemplate) error {
	return s.repo.Update(ctx, t)
}

// Delete removes a template owned by the user.
func (s *TemplateService) Delete(ctx context.Context, id, userID int) error {
	return s.repo.Delete(ctx, id, userID)
}

// CreateWorkoutFromTemplate builds a workout from a template.
func (s *TemplateService) CreateWorkoutFromTemplate(ctx context.Context, t *models.WorkoutTemplate) *models.Workout {
	w := &models.Workout{
		UserID: t.UserID,
		Title:  t.Name,
	}

	for _, te := range t.Exercises {
		w.Exercises = append(w.Exercises, models.WorkoutExercise{
			ExerciseID: te.ExerciseID,
			Sets:       te.Sets,
			Reps:       te.Reps,
			WeightKg:   te.WeightKg,
		})
	}

	return w
}
