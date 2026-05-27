package models

import "time"

// Workout represents a single training session.
type Workout struct {
	ID              int               `json:"id" db:"id"`
	UserID          int               `json:"user_id" db:"user_id"`
	Title           string            `json:"title" db:"title"`
	Date            time.Time         `json:"date" db:"date"`
	DurationMinutes int               `json:"duration_minutes" db:"duration_minutes"`
	Notes           string            `json:"notes" db:"notes"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
	Exercises       []WorkoutExercise `json:"exercises,omitempty" db:"-"`
}

// WorkoutExercise represents a specific exercise performed during a workout.
type WorkoutExercise struct {
	ID         int     `json:"id" db:"id"`
	WorkoutID  int     `json:"workout_id" db:"workout_id"`
	ExerciseID int     `json:"exercise_id" db:"exercise_id"`
	Sets       int     `json:"sets" db:"sets"`
	Reps       int     `json:"reps" db:"reps"`
	WeightKg   float64 `json:"weight_kg" db:"weight_kg"`
}
