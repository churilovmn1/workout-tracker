package service

import (
	"context"
	"errors"
	"time"

	"github.com/churilovmn1/workout-tracker/internal/models"
)

var ErrForbidden = errors.New("access denied")

type workoutRepository interface {
	Create(ctx context.Context, w *models.Workout) (int, error)
	GetByID(ctx context.Context, id int) (*models.Workout, error)
	ListByUser(ctx context.Context, userID int) ([]models.Workout, error)
	Update(ctx context.Context, w *models.Workout) error
	Delete(ctx context.Context, id, userID int) error
	GetPersonalRecords(ctx context.Context, userID int) ([]models.WorkoutExercise, error)
	GetWeeklyVolume(ctx context.Context, userID int) (float64, error)
}

// WorkoutService handles workout business logic.
type WorkoutService struct {
	repo workoutRepository
}

// NewWorkoutService creates a new WorkoutService.
func NewWorkoutService(repo workoutRepository) *WorkoutService {
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

// Update modifies a workout after verifying ownership.
func (s *WorkoutService) Update(ctx context.Context, w *models.Workout) error {
	existing, err := s.repo.GetByID(ctx, w.ID)
	if err != nil {
		return err
	}
	if existing.UserID != w.UserID {
		return ErrForbidden
	}
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

	dst := &models.Workout{
		UserID:    userID,
		Title:     source.Title,
		Date:      time.Now(),
		Notes:     source.Notes,
		Exercises: make([]models.WorkoutExercise, len(source.Exercises)),
	}

	for i, ex := range source.Exercises {
		dst.Exercises[i] = models.WorkoutExercise{
			ExerciseID: ex.ExerciseID,
			Sets:       ex.Sets,
			Reps:       ex.Reps,
			WeightKg:   ex.WeightKg,
		}
	}

	return s.repo.Create(ctx, dst)
}
