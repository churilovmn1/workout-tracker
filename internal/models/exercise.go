package models

// Exercise represents an exercise in the catalog.
type Exercise struct {
	ID          int    `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	MuscleGroup string `json:"muscle_group" db:"muscle_group"`
	Description string `json:"description" db:"description"`
}
