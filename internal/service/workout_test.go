package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/churilovmn1/workout-tracker/internal/models"
)

type mockWorkoutRepo struct {
	workouts map[int]*models.Workout
	nextID   int
}

func newMockWorkoutRepo() *mockWorkoutRepo {
	return &mockWorkoutRepo{workouts: make(map[int]*models.Workout), nextID: 1}
}

func (m *mockWorkoutRepo) Create(_ context.Context, w *models.Workout) (int, error) {
	w.ID = m.nextID
	m.nextID++
	clone := *w
	clone.Exercises = append([]models.WorkoutExercise(nil), w.Exercises...)
	m.workouts[clone.ID] = &clone
	return clone.ID, nil
}

func (m *mockWorkoutRepo) GetByID(_ context.Context, id int) (*models.Workout, error) {
	w, ok := m.workouts[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	clone := *w
	return &clone, nil
}

func (m *mockWorkoutRepo) ListByUser(_ context.Context, userID int) ([]models.Workout, error) {
	var result []models.Workout
	for _, w := range m.workouts {
		if w.UserID == userID {
			result = append(result, *w)
		}
	}
	return result, nil
}

func (m *mockWorkoutRepo) Update(_ context.Context, w *models.Workout) error {
	if _, ok := m.workouts[w.ID]; !ok {
		return fmt.Errorf("not found")
	}
	clone := *w
	m.workouts[w.ID] = &clone
	return nil
}

func (m *mockWorkoutRepo) Delete(_ context.Context, id, userID int) error {
	w, ok := m.workouts[id]
	if !ok || w.UserID != userID {
		return fmt.Errorf("not found")
	}
	delete(m.workouts, id)
	return nil
}

func (m *mockWorkoutRepo) GetPersonalRecords(_ context.Context, userID int) ([]models.WorkoutExercise, error) {
	return nil, nil
}

func (m *mockWorkoutRepo) GetWeeklyVolume(_ context.Context, userID int) (float64, error) {
	return 0, nil
}

func newTestWorkoutService() (*WorkoutService, *mockWorkoutRepo) {
	repo := newMockWorkoutRepo()
	return NewWorkoutService(repo), repo
}

func makeWorkout(userID int) *models.Workout {
	return &models.Workout{
		UserID: userID,
		Title:  "Test Workout",
		Date:   time.Now(),
	}
}

func TestWorkoutCreate(t *testing.T) {
	svc, _ := newTestWorkoutService()
	ctx := context.Background()

	id, err := svc.Create(ctx, makeWorkout(1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero workout id")
	}
}

func TestWorkoutGetByID_Owner(t *testing.T) {
	svc, _ := newTestWorkoutService()
	ctx := context.Background()

	id, _ := svc.Create(ctx, makeWorkout(1))

	w, err := svc.GetByID(ctx, id, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.ID != id {
		t.Errorf("expected id %d, got %d", id, w.ID)
	}
}

func TestWorkoutGetByID_NotOwner(t *testing.T) {
	svc, _ := newTestWorkoutService()
	ctx := context.Background()

	id, _ := svc.Create(ctx, makeWorkout(1))

	_, err := svc.GetByID(ctx, id, 2)
	if err == nil {
		t.Error("expected forbidden error, got nil")
	}
}

func TestWorkoutUpdate_Owner(t *testing.T) {
	svc, _ := newTestWorkoutService()
	ctx := context.Background()

	id, _ := svc.Create(ctx, makeWorkout(1))

	updated := makeWorkout(1)
	updated.ID = id
	updated.Title = "Updated"

	if err := svc.Update(ctx, updated); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	w, _ := svc.GetByID(ctx, id, 1)
	if w.Title != "Updated" {
		t.Errorf("expected title Updated, got %s", w.Title)
	}
}

func TestWorkoutUpdate_NotOwner(t *testing.T) {
	svc, _ := newTestWorkoutService()
	ctx := context.Background()

	id, _ := svc.Create(ctx, makeWorkout(1))

	attacker := makeWorkout(2)
	attacker.ID = id

	err := svc.Update(ctx, attacker)
	if err == nil {
		t.Error("expected forbidden error, got nil")
	}
}

func TestWorkoutDelete_Owner(t *testing.T) {
	svc, _ := newTestWorkoutService()
	ctx := context.Background()

	id, _ := svc.Create(ctx, makeWorkout(1))

	if err := svc.Delete(ctx, id, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := svc.GetByID(ctx, id, 1)
	if err == nil {
		t.Error("expected not-found error after delete, got nil")
	}
}

func TestWorkoutCopyWorkout_Owner(t *testing.T) {
	svc, _ := newTestWorkoutService()
	ctx := context.Background()

	src := makeWorkout(1)
	src.Exercises = []models.WorkoutExercise{
		{ExerciseID: 5, Sets: 3, Reps: 10, WeightKg: 60},
	}
	srcID, _ := svc.Create(ctx, src)

	copyID, err := svc.CopyWorkout(ctx, srcID, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if copyID == srcID {
		t.Error("copy should have a different id than source")
	}

	cp, _ := svc.GetByID(ctx, copyID, 1)
	if len(cp.Exercises) != 1 {
		t.Errorf("expected 1 exercise in copy, got %d", len(cp.Exercises))
	}
}

func TestWorkoutCopyWorkout_NotOwner(t *testing.T) {
	svc, _ := newTestWorkoutService()
	ctx := context.Background()

	srcID, _ := svc.Create(ctx, makeWorkout(1))

	_, err := svc.CopyWorkout(ctx, srcID, 2)
	if err == nil {
		t.Error("expected forbidden error, got nil")
	}
}
