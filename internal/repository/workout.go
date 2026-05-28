package repository

import (
	"context"
	"fmt"

	"github.com/churilovmn1/workout-tracker/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WorkoutRepository handles database operations for workouts.
type WorkoutRepository struct {
	pool *pgxpool.Pool
}

// NewWorkoutRepository creates a new WorkoutRepository.
func NewWorkoutRepository(pool *pgxpool.Pool) *WorkoutRepository {
	return &WorkoutRepository{pool: pool}
}

// Create inserts a workout with its exercises in a single transaction.
func (r *WorkoutRepository) Create(ctx context.Context, w *models.Workout) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var id int
	err = tx.QueryRow(ctx,
		`INSERT INTO workouts (user_id, title, date, duration_minutes, notes)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		w.UserID, w.Title, w.Date, w.DurationMinutes, w.Notes,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert workout: %w", err)
	}

	for _, ex := range w.Exercises {
		_, err = tx.Exec(ctx,
			`INSERT INTO workout_exercises (workout_id, exercise_id, sets, reps, weight_kg)
			 VALUES ($1, $2, $3, $4, $5)`,
			id, ex.ExerciseID, ex.Sets, ex.Reps, ex.WeightKg,
		)
		if err != nil {
			return 0, fmt.Errorf("insert workout exercise: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}
	return id, nil
}

// GetByID returns a workout with its exercises.
func (r *WorkoutRepository) GetByID(ctx context.Context, id int) (*models.Workout, error) {
	w := &models.Workout{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, title, date, duration_minutes, notes, trainer_comment, created_at
		 FROM workouts WHERE id = $1`, id,
	).Scan(&w.ID, &w.UserID, &w.Title, &w.Date, &w.DurationMinutes, &w.Notes, &w.TrainerComment, &w.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get workout: %w", err)
	}

	exercises, err := r.getExercises(ctx, id)
	if err != nil {
		return nil, err
	}
	w.Exercises = exercises

	return w, nil
}

// ListByUser returns all workouts for a user, ordered by date descending.
func (r *WorkoutRepository) ListByUser(ctx context.Context, userID int) ([]models.Workout, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, title, date, duration_minutes, notes, trainer_comment, created_at
		 FROM workouts WHERE user_id = $1
		 ORDER BY date DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list workouts: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Workout, error) {
		var w models.Workout
		err := row.Scan(&w.ID, &w.UserID, &w.Title, &w.Date, &w.DurationMinutes, &w.Notes, &w.TrainerComment, &w.CreatedAt)
		return w, err
	})
}

// Update modifies a workout and replaces its exercises.
func (r *WorkoutRepository) Update(ctx context.Context, w *models.Workout) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`UPDATE workouts SET title = $1, date = $2, duration_minutes = $3, notes = $4
		 WHERE id = $5 AND user_id = $6`,
		w.Title, w.Date, w.DurationMinutes, w.Notes, w.ID, w.UserID,
	)
	if err != nil {
		return fmt.Errorf("update workout: %w", err)
	}

	_, err = tx.Exec(ctx, `DELETE FROM workout_exercises WHERE workout_id = $1`, w.ID)
	if err != nil {
		return fmt.Errorf("delete old exercises: %w", err)
	}

	for _, ex := range w.Exercises {
		_, err = tx.Exec(ctx,
			`INSERT INTO workout_exercises (workout_id, exercise_id, sets, reps, weight_kg)
			 VALUES ($1, $2, $3, $4, $5)`,
			w.ID, ex.ExerciseID, ex.Sets, ex.Reps, ex.WeightKg,
		)
		if err != nil {
			return fmt.Errorf("insert workout exercise: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// Delete removes a workout by ID (cascade deletes exercises).
func (r *WorkoutRepository) Delete(ctx context.Context, id, userID int) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM workouts WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete workout: %w", err)
	}
	return nil
}

// SetTrainerComment sets the trainer's comment on any workout (admin operation).
func (r *WorkoutRepository) SetTrainerComment(ctx context.Context, id int, comment string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE workouts SET trainer_comment = $1 WHERE id = $2`, comment, id)
	if err != nil {
		return fmt.Errorf("set trainer comment: %w", err)
	}
	return nil
}

// GetPersonalRecords returns the max weight per exercise for a user.
func (r *WorkoutRepository) GetPersonalRecords(ctx context.Context, userID int) ([]models.WorkoutExercise, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT ON (we.exercise_id)
			we.id, we.workout_id, we.exercise_id, we.sets, we.reps, we.weight_kg
		 FROM workout_exercises we
		 JOIN workouts w ON w.id = we.workout_id
		 WHERE w.user_id = $1
		 ORDER BY we.exercise_id, we.weight_kg DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("get personal records: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.WorkoutExercise, error) {
		var we models.WorkoutExercise
		err := row.Scan(&we.ID, &we.WorkoutID, &we.ExerciseID, &we.Sets, &we.Reps, &we.WeightKg)
		return we, err
	})
}

// GetWeeklyVolume returns total volume (sets * reps * weight) for a user in the last 7 days.
func (r *WorkoutRepository) GetWeeklyVolume(ctx context.Context, userID int) (float64, error) {
	var volume float64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(we.sets * we.reps * we.weight_kg), 0)
		 FROM workout_exercises we
		 JOIN workouts w ON w.id = we.workout_id
		 WHERE w.user_id = $1 AND w.date >= CURRENT_DATE - INTERVAL '7 days'`,
		userID,
	).Scan(&volume)
	if err != nil {
		return 0, fmt.Errorf("get weekly volume: %w", err)
	}
	return volume, nil
}

func (r *WorkoutRepository) getExercises(ctx context.Context, workoutID int) ([]models.WorkoutExercise, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, workout_id, exercise_id, sets, reps, weight_kg
		 FROM workout_exercises WHERE workout_id = $1
		 ORDER BY id`, workoutID)
	if err != nil {
		return nil, fmt.Errorf("get workout exercises: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.WorkoutExercise, error) {
		var we models.WorkoutExercise
		err := row.Scan(&we.ID, &we.WorkoutID, &we.ExerciseID, &we.Sets, &we.Reps, &we.WeightKg)
		return we, err
	})
}
