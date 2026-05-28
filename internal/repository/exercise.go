package repository

import (
	"context"
	"fmt"

	"github.com/churilovmn1/workout-tracker/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ExerciseRepository handles database operations for exercises.
type ExerciseRepository struct {
	pool *pgxpool.Pool
}

// NewExerciseRepository creates a new ExerciseRepository.
func NewExerciseRepository(pool *pgxpool.Pool) *ExerciseRepository {
	return &ExerciseRepository{pool: pool}
}

// Create inserts a new exercise and returns its ID.
func (r *ExerciseRepository) Create(ctx context.Context, ex *models.Exercise) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx,
		`INSERT INTO exercises (name, muscle_group, description)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		ex.Name, ex.MuscleGroup, ex.Description,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create exercise: %w", err)
	}
	return id, nil
}

// GetByID returns an exercise by ID.
func (r *ExerciseRepository) GetByID(ctx context.Context, id int) (*models.Exercise, error) {
	ex := &models.Exercise{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, muscle_group, description
		 FROM exercises WHERE id = $1`, id,
	).Scan(&ex.ID, &ex.Name, &ex.MuscleGroup, &ex.Description)
	if err != nil {
		return nil, fmt.Errorf("get exercise by id: %w", err)
	}
	return ex, nil
}

// List returns exercises filtered by muscle_group and/or name search.
func (r *ExerciseRepository) List(ctx context.Context, muscleGroup, search string) ([]models.Exercise, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, muscle_group, description
		 FROM exercises
		 WHERE ($1 = '' OR muscle_group = $1)
		   AND ($2 = '' OR name ILIKE '%' || $2 || '%')
		 ORDER BY muscle_group, name`,
		muscleGroup, search)
	if err != nil {
		return nil, fmt.Errorf("list exercises: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Exercise, error) {
		var ex models.Exercise
		err := row.Scan(&ex.ID, &ex.Name, &ex.MuscleGroup, &ex.Description)
		return ex, err
	})
}

// Update modifies an existing exercise.
func (r *ExerciseRepository) Update(ctx context.Context, ex *models.Exercise) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE exercises SET name = $1, muscle_group = $2, description = $3
		 WHERE id = $4`,
		ex.Name, ex.MuscleGroup, ex.Description, ex.ID,
	)
	if err != nil {
		return fmt.Errorf("update exercise: %w", err)
	}
	return nil
}

// Delete removes an exercise by ID.
func (r *ExerciseRepository) Delete(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM exercises WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete exercise: %w", err)
	}
	return nil
}
