package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/churilovmn1/workout-tracker/internal/models"
	"github.com/churilovmn1/workout-tracker/internal/service"
	"github.com/go-chi/chi/v5"
)

// WorkoutHandler handles workout endpoints.
type WorkoutHandler struct {
	workoutService *service.WorkoutService
}

// NewWorkoutHandler creates a new WorkoutHandler.
func NewWorkoutHandler(workoutService *service.WorkoutService) *WorkoutHandler {
	return &WorkoutHandler{workoutService: workoutService}
}

type workoutExerciseRequest struct {
	ExerciseID int     `json:"exercise_id"`
	Sets       int     `json:"sets"`
	Reps       int     `json:"reps"`
	WeightKg   float64 `json:"weight_kg"`
}

type workoutRequest struct {
	Title           string                   `json:"title"`
	Date            string                   `json:"date"`
	DurationMinutes int                      `json:"duration_minutes"`
	Notes           string                   `json:"notes"`
	Exercises       []workoutExerciseRequest  `json:"exercises"`
}

type volumeResponse struct {
	WeeklyVolume float64 `json:"weekly_volume"`
}

// List returns all workouts for the authenticated user.
func (h *WorkoutHandler) List(w http.ResponseWriter, r *http.Request) {
	workouts, err := h.workoutService.ListByUser(r.Context(), getUserID(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list workouts")
		return
	}

	writeJSON(w, http.StatusOK, workouts)
}

// GetByID returns a workout by ID.
func (h *WorkoutHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid workout id")
		return
	}

	workout, err := h.workoutService.GetByID(r.Context(), id, getUserID(r))
	if err != nil {
		writeError(w, http.StatusNotFound, "workout not found")
		return
	}

	writeJSON(w, http.StatusOK, workout)
}

// Create adds a new workout.
func (h *WorkoutHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req workoutRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		date = time.Now()
	}

	workout := &models.Workout{
		UserID:          getUserID(r),
		Title:           req.Title,
		Date:            date,
		DurationMinutes: req.DurationMinutes,
		Notes:           req.Notes,
	}

	for _, e := range req.Exercises {
		workout.Exercises = append(workout.Exercises, models.WorkoutExercise{
			ExerciseID: e.ExerciseID,
			Sets:       e.Sets,
			Reps:       e.Reps,
			WeightKg:   e.WeightKg,
		})
	}

	id, err := h.workoutService.Create(r.Context(), workout)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create workout")
		return
	}

	workout.ID = id
	writeJSON(w, http.StatusCreated, workout)
}

// Update modifies an existing workout.
func (h *WorkoutHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid workout id")
		return
	}

	var req workoutRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		date = time.Now()
	}

	workout := &models.Workout{
		ID:              id,
		UserID:          getUserID(r),
		Title:           req.Title,
		Date:            date,
		DurationMinutes: req.DurationMinutes,
		Notes:           req.Notes,
	}

	for _, e := range req.Exercises {
		workout.Exercises = append(workout.Exercises, models.WorkoutExercise{
			ExerciseID: e.ExerciseID,
			Sets:       e.Sets,
			Reps:       e.Reps,
			WeightKg:   e.WeightKg,
		})
	}

	if err := h.workoutService.Update(r.Context(), workout); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update workout")
		return
	}

	writeJSON(w, http.StatusOK, workout)
}

// Delete removes a workout.
func (h *WorkoutHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid workout id")
		return
	}

	if err := h.workoutService.Delete(r.Context(), id, getUserID(r)); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete workout")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Copy creates a new workout based on an existing one.
func (h *WorkoutHandler) Copy(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid workout id")
		return
	}

	newID, err := h.workoutService.CopyWorkout(r.Context(), id, getUserID(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to copy workout")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]int{"id": newID})
}

// PersonalRecords returns best weights per exercise.
func (h *WorkoutHandler) PersonalRecords(w http.ResponseWriter, r *http.Request) {
	records, err := h.workoutService.GetPersonalRecords(r.Context(), getUserID(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get personal records")
		return
	}

	writeJSON(w, http.StatusOK, records)
}

// WeeklyVolume returns total volume for the last 7 days.
func (h *WorkoutHandler) WeeklyVolume(w http.ResponseWriter, r *http.Request) {
	volume, err := h.workoutService.GetWeeklyVolume(r.Context(), getUserID(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get weekly volume")
		return
	}

	writeJSON(w, http.StatusOK, volumeResponse{WeeklyVolume: volume})
}
