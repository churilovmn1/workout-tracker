package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/churilovmn1/workout-tracker/internal/models"
)

type mockUserRepo struct {
	byLogin map[string]*models.User
	nextID  int
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{byLogin: make(map[string]*models.User), nextID: 1}
}

func (m *mockUserRepo) Create(_ context.Context, user *models.User) (int, error) {
	if _, exists := m.byLogin[user.Login]; exists {
		return 0, fmt.Errorf("duplicate login")
	}
	user.ID = m.nextID
	m.nextID++
	clone := *user
	m.byLogin[user.Login] = &clone
	return user.ID, nil
}

func (m *mockUserRepo) GetByLogin(_ context.Context, login string) (*models.User, error) {
	u, ok := m.byLogin[login]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return u, nil
}

func newTestAuthService() *AuthService {
	return NewAuthService(newMockUserRepo(), "test-secret-key-32-bytes-long!!")
}

func TestRegister_Success(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()

	user, err := svc.Register(ctx, "alice", "alice@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID == 0 {
		t.Error("expected non-zero user ID")
	}
	if user.Login != "alice" {
		t.Errorf("expected login alice, got %s", user.Login)
	}
	if user.Role != models.RoleUser {
		t.Errorf("expected role user, got %s", user.Role)
	}
}

func TestRegister_DuplicateLogin(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()

	if _, err := svc.Register(ctx, "alice", "alice@example.com", "pass"); err != nil {
		t.Fatalf("first registration failed: %v", err)
	}
	if _, err := svc.Register(ctx, "alice", "alice2@example.com", "pass"); err == nil {
		t.Error("expected error on duplicate login, got nil")
	}
}

func TestLogin_ValidCredentials(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()

	if _, err := svc.Register(ctx, "bob", "bob@example.com", "secret"); err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	token, err := svc.Login(ctx, "bob", "secret")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()

	if _, err := svc.Register(ctx, "carol", "carol@example.com", "correct"); err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	_, err := svc.Login(ctx, "carol", "wrong")
	if err == nil {
		t.Error("expected error on wrong password, got nil")
	}
}

func TestLogin_UnknownUser(t *testing.T) {
	svc := newTestAuthService()
	_, err := svc.Login(context.Background(), "nobody", "pass")
	if err == nil {
		t.Error("expected error for unknown user, got nil")
	}
}

func TestParseToken_RoundTrip(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()

	user, _ := svc.Register(ctx, "dave", "dave@example.com", "pass")

	token, err := svc.Login(ctx, "dave", "pass")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	claims, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("parse token failed: %v", err)
	}
	if claims.UserID != user.ID {
		t.Errorf("expected user id %d, got %d", user.ID, claims.UserID)
	}
	if claims.Role != models.RoleUser {
		t.Errorf("expected role user, got %s", claims.Role)
	}
}

func TestParseToken_InvalidToken(t *testing.T) {
	svc := newTestAuthService()
	_, err := svc.ParseToken("not.a.token")
	if err == nil {
		t.Error("expected error on invalid token, got nil")
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	svc1 := NewAuthService(newMockUserRepo(), "secret-one")
	svc2 := NewAuthService(newMockUserRepo(), "secret-two")

	tok, err := svc1.GenerateToken(&models.User{ID: 1, Role: models.RoleUser})
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	_, err = svc2.ParseToken(tok)
	if err == nil {
		t.Error("expected error when parsing token with wrong secret, got nil")
	}
}
