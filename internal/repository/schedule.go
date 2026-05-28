package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/churilovmn1/workout-tracker/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ScheduleRepository handles database operations for the trainer schedule.
type ScheduleRepository struct {
	pool *pgxpool.Pool
}

// NewScheduleRepository creates a new ScheduleRepository.
func NewScheduleRepository(pool *pgxpool.Pool) *ScheduleRepository {
	return &ScheduleRepository{pool: pool}
}

// Create inserts a new schedule entry and returns its ID.
func (r *ScheduleRepository) Create(ctx context.Context, e *models.ScheduleEntry) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx,
		`INSERT INTO schedule (trainer_id, client_id, title, scheduled_at, duration_minutes, status, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		e.TrainerID, e.ClientID, e.Title, e.ScheduledAt, e.DurationMinutes, e.Status, e.Notes,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create schedule entry: %w", err)
	}
	return id, nil
}

// GetByID returns a schedule entry with client login.
func (r *ScheduleRepository) GetByID(ctx context.Context, id int) (*models.ScheduleEntry, error) {
	e := &models.ScheduleEntry{}
	err := r.pool.QueryRow(ctx,
		`SELECT s.id, s.trainer_id, s.client_id, s.title, s.scheduled_at,
		        s.duration_minutes, s.status, s.notes, s.created_at, u.login
		 FROM schedule s
		 JOIN users u ON u.id = s.client_id
		 WHERE s.id = $1`, id,
	).Scan(&e.ID, &e.TrainerID, &e.ClientID, &e.Title, &e.ScheduledAt,
		&e.DurationMinutes, &e.Status, &e.Notes, &e.CreatedAt, &e.ClientLogin)
	if err != nil {
		return nil, fmt.Errorf("get schedule entry: %w", err)
	}
	return e, nil
}

// ListByTrainerWeek returns entries for a trainer within [start, end).
func (r *ScheduleRepository) ListByTrainerWeek(ctx context.Context, trainerID int, start, end time.Time) ([]models.ScheduleEntry, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT s.id, s.trainer_id, s.client_id, s.title, s.scheduled_at,
		        s.duration_minutes, s.status, s.notes, s.created_at, u.login
		 FROM schedule s
		 JOIN users u ON u.id = s.client_id
		 WHERE s.trainer_id = $1
		   AND s.scheduled_at >= $2
		   AND s.scheduled_at < $3
		 ORDER BY s.scheduled_at`,
		trainerID, start, end)
	if err != nil {
		return nil, fmt.Errorf("list schedule: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.ScheduleEntry, error) {
		var e models.ScheduleEntry
		err := row.Scan(&e.ID, &e.TrainerID, &e.ClientID, &e.Title, &e.ScheduledAt,
			&e.DurationMinutes, &e.Status, &e.Notes, &e.CreatedAt, &e.ClientLogin)
		return e, err
	})
}

// Update modifies a schedule entry owned by the trainer.
func (r *ScheduleRepository) Update(ctx context.Context, e *models.ScheduleEntry) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE schedule
		 SET client_id = $1, title = $2, scheduled_at = $3,
		     duration_minutes = $4, status = $5, notes = $6
		 WHERE id = $7 AND trainer_id = $8`,
		e.ClientID, e.Title, e.ScheduledAt, e.DurationMinutes, e.Status, e.Notes,
		e.ID, e.TrainerID,
	)
	if err != nil {
		return fmt.Errorf("update schedule entry: %w", err)
	}
	return nil
}

// Delete removes a schedule entry owned by the trainer.
func (r *ScheduleRepository) Delete(ctx context.Context, id, trainerID int) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM schedule WHERE id = $1 AND trainer_id = $2`, id, trainerID)
	if err != nil {
		return fmt.Errorf("delete schedule entry: %w", err)
	}
	return nil
}
