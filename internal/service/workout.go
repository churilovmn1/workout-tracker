package service

import (
	"context"
	"errors"

	"github.com/churilovmn1/workout-tracker/internal/models"
	"github.com/churilovmn1/workout-tracker/internal/repository"
)

var ErrForbidden = errors.New("access denied")

// WorkoutService handles workout business logic.
type WorkoutService struct {
	repo *repository.WorkoutRepository
}

// NewWorkoutService creates a new WorkoutService.
func NewWorkoutService(repo *repository.WorkoutRepository) *WorkoutService {
	return &WorkoutService{repo: repo}
}

// Create adds a new workout for the user.
func (s *WorkoutService) Create(ctx context.Context, w *models.Workout) (int, error) {
	return s.repo.Create(ctx, w)
}

// GetByID returns a workout if it belongs to the user.
func (s *WorkoutService) GetByID(ctx context.Context, id, userID int) (*models.Workout, error) {
	w, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if w.UserID != userID {
		return nil, ErrForbidden
	}
	return w, nil
}

// ListByUser returns all workouts for a user.
func (s *WorkoutService) ListByUser(ctx context.Context, userID int) ([]models.Workout, error) {
	return s.repo.ListByUser(ctx, userID)
}

// Update modifies a workout owned by the user.
func (s *WorkoutService) Update(ctx context.Context, w *models.Workout) error {
	return s.repo.Update(ctx, w)
}

// Delete removes a workout owned by the user.
func (s *WorkoutService) Delete(ctx context.Context, id, userID int) error {
	return s.repo.Delete(ctx, id, userID)
}

// GetPersonalRecords returns best weight per exercise for the user.
func (s *WorkoutService) GetPersonalRecords(ctx context.Context, userID int) ([]models.WorkoutExercise, error) {
	return s.repo.GetPersonalRecords(ctx, userID)
}

// GetWeeklyVolume returns total training volume for the last 7 days.
func (s *WorkoutService) GetWeeklyVolume(ctx context.Context, userID int) (float64, error) {
	return s.repo.GetWeeklyVolume(ctx, userID)
}

// CopyWorkout creates a new workout based on an existing one.
func (s *WorkoutService) CopyWorkout(ctx context.Context, sourceID, userID int) (int, error) {
	source, err := s.repo.GetByID(ctx, sourceID)
	if err != nil {
		return 0, err
	}
	if source.UserID != userID {
		return 0, ErrForbidden
	}

	copy := &models.Workout{
		UserID:    userID,
		Title:     source.Title,
		Date:      source.Date,
		Notes:     source.Notes,
		Exercises: source.Exercises,
	}

	for i := range copy.Exercises {
		copy.Exercises[i].ID = 0
		copy.Exercises[i].WorkoutID = 0
	}

	return s.repo.Create(ctx, copy)
}
