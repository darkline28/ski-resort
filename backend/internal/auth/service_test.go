package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/darkline28/ski-resort/backend/internal/users"
)

type FakeRepository struct {
	CreateCalled bool
}

func (f *FakeRepository) Create(ctx context.Context, user users.User) (int, error) {
	f.CreateCalled = true
	return 99, nil
}

func TestRegister_InvalidEmail(t *testing.T) {
	repo := FakeRepository{}
	service := NewService(&repo)

	input := RegisterInput{
		Name:     "Gosha",
		Email:    "email.org",
		Phone:    "+78945678491",
		Password: "testpassword",
	}
	user, err := service.Register(context.Background(), input)
	if user != (users.User{}) {
		t.Errorf("Expected no user to be created, but got %v", user)
	}
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("Expected ErrInvalidInput, got %v", err)
	}
	if repo.CreateCalled {
		t.Errorf("repo.Create не должен вызываться при некорректном email")
	}

}
