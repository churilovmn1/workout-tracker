package models

import "time"

const (
	ScheduleStatusPlanned   = "planned"
	ScheduleStatusCompleted = "completed"
	ScheduleStatusCancelled = "cancelled"
)

// ScheduleEntry is a trainer-scheduled session with a client.
type ScheduleEntry struct {
	ID              int       `json:"id"`
	TrainerID       int       `json:"trainer_id"`
	ClientID        int       `json:"client_id"`
	Title           string    `json:"title"`
	ScheduledAt     time.Time `json:"scheduled_at"`
	DurationMinutes int       `json:"duration_minutes"`
	Status          string    `json:"status"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
	ClientLogin     string    `json:"client_login"` // populated via JOIN
}
