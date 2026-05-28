package service

import (
	"context"
	"time"

	"github.com/churilovmn1/workout-tracker/internal/models"
)

type templateRepository interface {
	Create(ctx context.Context, t *models.WorkoutTemplate) (int, error)
	GetByID(ctx context.Context, id int) (*models.WorkoutTemplate, error)
	ListByUser(ctx context.Context, userID int) ([]models.WorkoutTemplate, error)
	Update(ctx context.Context, t *models.WorkoutTemplate) error
	Delete(ctx context.Context, id, userID int) error
}

// TemplateService handles workout template business logic.
type TemplateService struct {
	repo templateRepository
}

// NewTemplateService creates a new TemplateService.
func NewTemplateService(repo templateRepository) *TemplateService {
	return &TemplateService{repo: repo}
}

// Create adds a new workout template.
func (s *TemplateService) Create(ctx context.Context, t *models.WorkoutTemplate) (int, error) {
	return s.repo.Create(ctx, t)
}

// GetByID returns a template if the user has access (owner or public).
func (s *TemplateService) GetByID(ctx context.Context, id, userID int) (*models.WorkoutTemplate, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t.UserID != userID && !t.IsPublic {
		return nil, ErrForbidden
	}
	return t, nil
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

// CreateWorkoutFromTemplate builds a workout struct from a template for the given user.
func (s *TemplateService) CreateWorkoutFromTemplate(t *models.WorkoutTemplate, userID int) *models.Workout {
	w := &models.Workout{
		UserID: userID,
		Title:  t.Name,
		Date:   time.Now(),
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
