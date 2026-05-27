package models

// WorkoutTemplate represents a reusable workout plan.
type WorkoutTemplate struct {
	ID        int                `json:"id" db:"id"`
	UserID    int                `json:"user_id" db:"user_id"`
	Name      string             `json:"name" db:"name"`
	IsPublic  bool               `json:"is_public" db:"is_public"`
	Exercises []TemplateExercise `json:"exercises,omitempty" db:"-"`
}

// TemplateExercise represents an exercise within a template.
type TemplateExercise struct {
	ID         int     `json:"id" db:"id"`
	TemplateID int     `json:"template_id" db:"template_id"`
	ExerciseID int     `json:"exercise_id" db:"exercise_id"`
	Sets       int     `json:"sets" db:"sets"`
	Reps       int     `json:"reps" db:"reps"`
	WeightKg   float64 `json:"weight_kg" db:"weight_kg"`
}
