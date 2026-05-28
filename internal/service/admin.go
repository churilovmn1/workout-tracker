package service

import (
	"context"
	"time"

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

type scheduleRepository interface {
	Create(ctx context.Context, e *models.ScheduleEntry) (int, error)
	GetByID(ctx context.Context, id int) (*models.ScheduleEntry, error)
	ListByTrainerWeek(ctx context.Context, trainerID int, start, end time.Time) ([]models.ScheduleEntry, error)
	Update(ctx context.Context, e *models.ScheduleEntry) error
	Delete(ctx context.Context, id, trainerID int) error
}

// AdminService provides trainer/admin operations.
type AdminService struct {
	userRepo     adminUserRepository
	workoutRepo  adminWorkoutRepository
	scheduleRepo scheduleRepository
}

// NewAdminService creates a new AdminService.
func NewAdminService(
	userRepo adminUserRepository,
	workoutRepo adminWorkoutRepository,
	scheduleRepo scheduleRepository,
) *AdminService {
	return &AdminService{
		userRepo:     userRepo,
		workoutRepo:  workoutRepo,
		scheduleRepo: scheduleRepo,
	}
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

// ListScheduleWeek returns schedule entries for a trainer in [start, end).
func (s *AdminService) ListScheduleWeek(ctx context.Context, trainerID int, start, end time.Time) ([]models.ScheduleEntry, error) {
	return s.scheduleRepo.ListByTrainerWeek(ctx, trainerID, start, end)
}

// GetScheduleEntry returns a single schedule entry by ID.
func (s *AdminService) GetScheduleEntry(ctx context.Context, id int) (*models.ScheduleEntry, error) {
	return s.scheduleRepo.GetByID(ctx, id)
}

// CreateScheduleEntry adds a new schedule entry.
func (s *AdminService) CreateScheduleEntry(ctx context.Context, e *models.ScheduleEntry) (int, error) {
	if e.Status == "" {
		e.Status = models.ScheduleStatusPlanned
	}
	return s.scheduleRepo.Create(ctx, e)
}

// UpdateScheduleEntry updates an existing schedule entry.
func (s *AdminService) UpdateScheduleEntry(ctx context.Context, e *models.ScheduleEntry) error {
	return s.scheduleRepo.Update(ctx, e)
}

// DeleteScheduleEntry removes a schedule entry owned by the trainer.
func (s *AdminService) DeleteScheduleEntry(ctx context.Context, id, trainerID int) error {
	return s.scheduleRepo.Delete(ctx, id, trainerID)
}
