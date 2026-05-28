package service

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/churilovmn1/workout-tracker/internal/models"
)

type mockExerciseRepo struct {
	exercises map[int]*models.Exercise
	nextID    int
}

func newMockExerciseRepo() *mockExerciseRepo {
	return &mockExerciseRepo{exercises: make(map[int]*models.Exercise), nextID: 1}
}

func (m *mockExerciseRepo) Create(_ context.Context, ex *models.Exercise) (int, error) {
	ex.ID = m.nextID
	m.nextID++
	clone := *ex
	m.exercises[clone.ID] = &clone
	return clone.ID, nil
}

func (m *mockExerciseRepo) GetByID(_ context.Context, id int) (*models.Exercise, error) {
	ex, ok := m.exercises[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	clone := *ex
	return &clone, nil
}

func (m *mockExerciseRepo) List(_ context.Context, muscleGroup, search string) ([]models.Exercise, error) {
	var result []models.Exercise
	for _, ex := range m.exercises {
		if muscleGroup != "" && ex.MuscleGroup != muscleGroup {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(ex.Name), strings.ToLower(search)) {
			continue
		}
		result = append(result, *ex)
	}
	return result, nil
}

func (m *mockExerciseRepo) Update(_ context.Context, ex *models.Exercise) error {
	if _, ok := m.exercises[ex.ID]; !ok {
		return fmt.Errorf("not found")
	}
	clone := *ex
	m.exercises[ex.ID] = &clone
	return nil
}

func (m *mockExerciseRepo) Delete(_ context.Context, id int) error {
	if _, ok := m.exercises[id]; !ok {
		return fmt.Errorf("not found")
	}
	delete(m.exercises, id)
	return nil
}

func newTestExerciseService() *ExerciseService {
	return NewExerciseService(newMockExerciseRepo())
}

func TestExerciseCreate(t *testing.T) {
	svc := newTestExerciseService()
	ctx := context.Background()

	id, err := svc.Create(ctx, &models.Exercise{Name: "Bench Press", MuscleGroup: "chest"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero exercise id")
	}
}

func TestExerciseGetByID_Found(t *testing.T) {
	svc := newTestExerciseService()
	ctx := context.Background()

	id, _ := svc.Create(ctx, &models.Exercise{Name: "Squat", MuscleGroup: "legs"})

	ex, err := svc.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ex.Name != "Squat" {
		t.Errorf("expected name Squat, got %s", ex.Name)
	}
}

func TestExerciseGetByID_NotFound(t *testing.T) {
	svc := newTestExerciseService()
	_, err := svc.GetByID(context.Background(), 999)
	if err == nil {
		t.Error("expected error for missing exercise, got nil")
	}
}

func TestExerciseList_All(t *testing.T) {
	svc := newTestExerciseService()
	ctx := context.Background()

	svc.Create(ctx, &models.Exercise{Name: "Push-up", MuscleGroup: "chest"})
	svc.Create(ctx, &models.Exercise{Name: "Pull-up", MuscleGroup: "back"})
	svc.Create(ctx, &models.Exercise{Name: "Dip", MuscleGroup: "chest"})

	all, err := svc.List(ctx, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 exercises, got %d", len(all))
	}
}

func TestExerciseList_FilteredByMuscleGroup(t *testing.T) {
	svc := newTestExerciseService()
	ctx := context.Background()

	svc.Create(ctx, &models.Exercise{Name: "Push-up", MuscleGroup: "chest"})
	svc.Create(ctx, &models.Exercise{Name: "Pull-up", MuscleGroup: "back"})
	svc.Create(ctx, &models.Exercise{Name: "Dip", MuscleGroup: "chest"})

	chest, err := svc.List(ctx, "chest", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chest) != 2 {
		t.Errorf("expected 2 chest exercises, got %d", len(chest))
	}
}

func TestExerciseUpdate(t *testing.T) {
	svc := newTestExerciseService()
	ctx := context.Background()

	id, _ := svc.Create(ctx, &models.Exercise{Name: "Deadlift", MuscleGroup: "back"})

	err := svc.Update(ctx, &models.Exercise{ID: id, Name: "Romanian Deadlift", MuscleGroup: "hamstrings"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ex, _ := svc.GetByID(ctx, id)
	if ex.Name != "Romanian Deadlift" {
		t.Errorf("expected updated name, got %s", ex.Name)
	}
}

func TestExerciseDelete(t *testing.T) {
	svc := newTestExerciseService()
	ctx := context.Background()

	id, _ := svc.Create(ctx, &models.Exercise{Name: "Curl", MuscleGroup: "biceps"})

	if err := svc.Delete(ctx, id); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := svc.GetByID(ctx, id)
	if err == nil {
		t.Error("expected not-found error after delete, got nil")
	}
}
