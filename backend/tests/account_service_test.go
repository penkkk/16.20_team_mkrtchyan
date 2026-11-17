package tests

import (
	"context"
	"errors"
	"testing"

	authrepo "opd/internal/repository/auth"
	authservice "opd/internal/service/auth"

	"golang.org/x/crypto/bcrypt"
)

type accountRepository struct {
	*authrepo.AuthPostgresRepository
	passwordLookupErr error
	createPasswordErr error
	createdPassword   authrepo.CreateUserPasswordInput
	user              authrepo.User
	userErr           error
	conflicts         []string
}

func (r *accountRepository) GetUserPasswordByID(_ context.Context, _ authrepo.GetPasswordByUserIDInput) error {
	return r.passwordLookupErr
}

func (r *accountRepository) CreateUserPassword(_ context.Context, input authrepo.CreateUserPasswordInput) (authrepo.UserPassword, error) {
	r.createdPassword = input
	return authrepo.UserPassword{}, r.createPasswordErr
}

func (r *accountRepository) GetUserByID(_ context.Context, _ authrepo.GetUserByIDInput) (authrepo.User, error) {
	return r.user, r.userErr
}

func (r *accountRepository) FindUserConflicts(_ context.Context, _ authrepo.FindUserConflictsInput) ([]string, error) {
	return r.conflicts, nil
}

func TestAddPasswordService(t *testing.T) {
	ctx := context.Background()
	repo := &accountRepository{passwordLookupErr: authrepo.ErrNotFound}
	service := authservice.NewAuthService(repo, nil, nil, nil, authservice.OAuthConfig{})
	if err := service.AddPassword(ctx, authservice.AddPasswordInput{UserID: "user-1", Password: "new-secret"}); err != nil {
		t.Fatal(err)
	}
	if repo.createdPassword.UserID != "user-1" || bcrypt.CompareHashAndPassword([]byte(repo.createdPassword.PasswordHash), []byte("new-secret")) != nil {
		t.Fatal("password was not hashed and stored for the user")
	}
	for _, tc := range []struct {
		name      string
		lookupErr error
		createErr error
	}{
		{"already set", nil, nil},
		{"concurrent insert", authrepo.ErrNotFound, authrepo.NewUniqueConstraintError("user_id")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &accountRepository{passwordLookupErr: tc.lookupErr, createPasswordErr: tc.createErr}
			service := authservice.NewAuthService(repo, nil, nil, nil, authservice.OAuthConfig{})
			if err := service.AddPassword(ctx, authservice.AddPasswordInput{UserID: "user-1", Password: "secret"}); !errors.Is(err, authservice.ErrPasswordAlreadySet) {
				t.Fatalf("expected duplicate-password error, got %v", err)
			}
		})
	}
}

func TestMeService(t *testing.T) {
	repo := &accountRepository{user: authrepo.User{ID: "user-1", Username: "alice", Email: "alice@example.com"}}
	service := authservice.NewAuthService(repo, nil, nil, nil, authservice.OAuthConfig{})
	user, err := service.Me(context.Background(), authservice.MeInput{UserID: "user-1"})
	if err != nil || user.ID != "user-1" || user.Username != "alice" || user.Email != "alice@example.com" {
		t.Fatalf("unexpected profile: %+v, err=%v", user, err)
	}
	repo.userErr = authrepo.ErrNotFound
	if _, err := service.Me(context.Background(), authservice.MeInput{UserID: "missing"}); !errors.Is(err, authservice.ErrUserNotFound) {
		t.Fatalf("expected user-not-found error, got %v", err)
	}
}

func TestRegisterServiceRejectsConflictingFields(t *testing.T) {
	repo := &accountRepository{conflicts: []string{"username", "email"}}
	service := authservice.NewAuthService(repo, nil, nil, nil, authservice.OAuthConfig{})
	_, err := service.Register(context.Background(), authservice.RegisterInput{Username: "alice", Email: "alice@example.com"})
	var conflict *authservice.FieldConflictError
	if !errors.As(err, &conflict) || len(conflict.Fields) != 2 {
		t.Fatalf("expected username and email conflicts, got %v", err)
	}
}
