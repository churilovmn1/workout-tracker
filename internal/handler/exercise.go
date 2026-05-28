package handler

import (
	"net/http"
	"strconv"

	"github.com/churilovmn1/workout-tracker/internal/models"
	"github.com/churilovmn1/workout-tracker/internal/service"
	"github.com/go-chi/chi/v5"
)

// ExerciseHandler handles exercise catalog endpoints.
type ExerciseHandler struct {
	exerciseService *service.ExerciseService
}

// NewExerciseHandler creates a new ExerciseHandler.
func NewExerciseHandler(exerciseService *service.ExerciseService) *ExerciseHandler {
	return &ExerciseHandler{exerciseService: exerciseService}
}

type exerciseRequest struct {
	Name        string `json:"name"`
	MuscleGroup string `json:"muscle_group"`
	Description string `json:"description"`
}

// List returns exercises filtered by muscle_group and/or search query.
func (h *ExerciseHandler) List(w http.ResponseWriter, r *http.Request) {
	muscleGroup := r.URL.Query().Get("muscle_group")
	search := r.URL.Query().Get("search")

	exercises, err := h.exerciseService.List(r.Context(), muscleGroup, search)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list exercises")
		return
	}

	writeJSON(w, http.StatusOK, exercises)
}

// GetByID returns an exercise by ID.
func (h *ExerciseHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid exercise id")
		return
	}

	exercise, err := h.exerciseService.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "exercise not found")
		return
	}

	writeJSON(w, http.StatusOK, exercise)
}

// Create adds a new exercise (admin only).
func (h *ExerciseHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req exerciseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.MuscleGroup == "" {
		writeError(w, http.StatusBadRequest, "name and muscle_group are required")
		return
	}

	ex := &models.Exercise{
		Name:        req.Name,
		MuscleGroup: req.MuscleGroup,
		Description: req.Description,
	}

	id, err := h.exerciseService.Create(r.Context(), ex)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create exercise")
		return
	}

	ex.ID = id
	writeJSON(w, http.StatusCreated, ex)
}

// Update modifies an exercise (admin only).
func (h *ExerciseHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid exercise id")
		return
	}

	var req exerciseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ex := &models.Exercise{
		ID:          id,
		Name:        req.Name,
		MuscleGroup: req.MuscleGroup,
		Description: req.Description,
	}

	if err := h.exerciseService.Update(r.Context(), ex); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update exercise")
		return
	}

	writeJSON(w, http.StatusOK, ex)
}

// Delete removes an exercise (admin only).
func (h *ExerciseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid exercise id")
		return
	}

	if err := h.exerciseService.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete exercise")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
