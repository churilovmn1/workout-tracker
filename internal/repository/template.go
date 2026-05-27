package repository

import (
	"context"
	"fmt"

	"github.com/churilovmn1/workout-tracker/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TemplateRepository handles database operations for workout templates.
type TemplateRepository struct {
	pool *pgxpool.Pool
}

// NewTemplateRepository creates a new TemplateRepository.
func NewTemplateRepository(pool *pgxpool.Pool) *TemplateRepository {
	return &TemplateRepository{pool: pool}
}

// Create inserts a template with its exercises.
func (r *TemplateRepository) Create(ctx context.Context, t *models.WorkoutTemplate) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var id int
	err = tx.QueryRow(ctx,
		`INSERT INTO workout_templates (user_id, name, is_public)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		t.UserID, t.Name, t.IsPublic,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert template: %w", err)
	}

	for _, ex := range t.Exercises {
		_, err = tx.Exec(ctx,
			`INSERT INTO template_exercises (template_id, exercise_id, sets, reps, weight_kg)
			 VALUES ($1, $2, $3, $4, $5)`,
			id, ex.ExerciseID, ex.Sets, ex.Reps, ex.WeightKg,
		)
		if err != nil {
			return 0, fmt.Errorf("insert template exercise: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}
	return id, nil
}

// GetByID returns a template with its exercises.
func (r *TemplateRepository) GetByID(ctx context.Context, id int) (*models.WorkoutTemplate, error) {
	t := &models.WorkoutTemplate{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, name, is_public
		 FROM workout_templates WHERE id = $1`, id,
	).Scan(&t.ID, &t.UserID, &t.Name, &t.IsPublic)
	if err != nil {
		return nil, fmt.Errorf("get template: %w", err)
	}

	exercises, err := r.getExercises(ctx, id)
	if err != nil {
		return nil, err
	}
	t.Exercises = exercises

	return t, nil
}

// ListByUser returns user's own templates plus public templates from others.
func (r *TemplateRepository) ListByUser(ctx context.Context, userID int) ([]models.WorkoutTemplate, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, name, is_public
		 FROM workout_templates
		 WHERE user_id = $1 OR is_public = TRUE
		 ORDER BY name`, userID)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.WorkoutTemplate, error) {
		var t models.WorkoutTemplate
		err := row.Scan(&t.ID, &t.UserID, &t.Name, &t.IsPublic)
		return t, err
	})
}

// Update modifies a template and replaces its exercises.
func (r *TemplateRepository) Update(ctx context.Context, t *models.WorkoutTemplate) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`UPDATE workout_templates SET name = $1, is_public = $2
		 WHERE id = $3 AND user_id = $4`,
		t.Name, t.IsPublic, t.ID, t.UserID,
	)
	if err != nil {
		return fmt.Errorf("update template: %w", err)
	}

	_, err = tx.Exec(ctx, `DELETE FROM template_exercises WHERE template_id = $1`, t.ID)
	if err != nil {
		return fmt.Errorf("delete old exercises: %w", err)
	}

	for _, ex := range t.Exercises {
		_, err = tx.Exec(ctx,
			`INSERT INTO template_exercises (template_id, exercise_id, sets, reps, weight_kg)
			 VALUES ($1, $2, $3, $4, $5)`,
			t.ID, ex.ExerciseID, ex.Sets, ex.Reps, ex.WeightKg,
		)
		if err != nil {
			return fmt.Errorf("insert template exercise: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// Delete removes a template by ID.
func (r *TemplateRepository) Delete(ctx context.Context, id, userID int) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM workout_templates WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete template: %w", err)
	}
	return nil
}

func (r *TemplateRepository) getExercises(ctx context.Context, templateID int) ([]models.TemplateExercise, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, template_id, exercise_id, sets, reps, weight_kg
		 FROM template_exercises WHERE template_id = $1
		 ORDER BY id`, templateID)
	if err != nil {
		return nil, fmt.Errorf("get template exercises: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.TemplateExercise, error) {
		var te models.TemplateExercise
		err := row.Scan(&te.ID, &te.TemplateID, &te.ExerciseID, &te.Sets, &te.Reps, &te.WeightKg)
		return te, err
	})
}
