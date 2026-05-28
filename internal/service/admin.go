package service

import (
	"context"

	"github.com/churilovmn1/workout-tracker/internal/models"
)

type adminUserRepository interface {
	List(ctx context.Context) ([]models.User, error)
}

type adminWorkoutRepository interface {
	ListByUser(ctx context.Context, userID int) ([]models.Workout, error)
	Create(ctx context.Context, w *models.Workout) (int, error)
	SetTrainerComment(ctx context.Context, id int, comment string) error
}

// AdminService provides trainer/admin operations.
type AdminService struct {
	userRepo    adminUserRepository
	workoutRepo adminWorkoutRepository
}

// NewAdminService creates a new AdminService.
func NewAdminService(userRepo adminUserRepository, workoutRepo adminWorkoutRepository) *AdminService {
	return &AdminService{userRepo: userRepo, workoutRepo: workoutRepo}
}

// ListUsers returns all registered users.
func (s *AdminService) ListUsers(ctx context.Context) ([]models.User, error) {
	return s.userRepo.List(ctx)
}

// ListUserWorkouts returns all workouts for the given user.
func (s *AdminService) ListUserWorkouts(ctx context.Context, userID int) ([]models.Workout, error) {
	return s.workoutRepo.ListByUser(ctx, userID)
}

// SetTrainerComment sets a trainer comment on any workout.
func (s *AdminService) SetTrainerComment(ctx context.Context, workoutID int, comment string) error {
	return s.workoutRepo.SetTrainerComment(ctx, workoutID, comment)
}

// CreateWorkoutForUser creates a workout on behalf of the given user.
func (s *AdminService) CreateWorkoutForUser(ctx context.Context, w *models.Workout) (int, error) {
	return s.workoutRepo.Create(ctx, w)
}
