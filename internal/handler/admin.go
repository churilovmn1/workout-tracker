package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/churilovmn1/workout-tracker/internal/models"
	"github.com/churilovmn1/workout-tracker/internal/service"
	"github.com/go-chi/chi/v5"
)

// AdminHandler handles trainer/admin endpoints.
type AdminHandler struct {
	adminService *service.AdminService
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

type commentRequest struct {
	Comment string `json:"comment"`
}

// ListUsers returns all registered users.
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.adminService.ListUsers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	writeJSON(w, http.StatusOK, users)
}

// ListUserWorkouts returns all workouts for a specific user.
func (h *AdminHandler) ListUserWorkouts(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	workouts, err := h.adminService.ListUserWorkouts(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list workouts")
		return
	}
	writeJSON(w, http.StatusOK, workouts)
}

// SetComment sets a trainer comment on a workout.
func (h *AdminHandler) SetComment(w http.ResponseWriter, r *http.Request) {
	workoutID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid workout id")
		return
	}

	var req commentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.adminService.SetTrainerComment(r.Context(), workoutID, req.Comment); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to set comment")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// CreateWorkoutForUser creates a workout on behalf of a user.
func (h *AdminHandler) CreateWorkoutForUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

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
		UserID:          userID,
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

	id, err := h.adminService.CreateWorkoutForUser(r.Context(), workout)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create workout")
		return
	}

	workout.ID = id
	writeJSON(w, http.StatusCreated, workout)
}
