package tests

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	authrepo "opd/internal/repository/auth"
	authservice "opd/internal/service/auth"
	"opd/internal/service/token"

	"golang.org/x/crypto/bcrypt"
)

type loginRepository struct {
	*authrepo.AuthPostgresRepository
	credentials  authrepo.UserCredentials
	lookupErr    error
	login        string
	revokedUser  string
	revokedAgent *string
	session      authrepo.CreateRefreshSessionInput
	created      bool
}

func (r *loginRepository) GetUserCredentialsByLogin(_ context.Context, login string) (authrepo.UserCredentials, error) {
	r.login = login
	return r.credentials, r.lookupErr
}

func (r *loginRepository) RevokeActiveRefreshSessionsByUserAgent(_ context.Context, userID string, agent *string) error {
	r.revokedUser, r.revokedAgent = userID, agent
	return nil
}

func (r *loginRepository) CreateRefreshSession(_ context.Context, input authrepo.CreateRefreshSessionInput) (authrepo.RefreshSession, error) {
	r.session, r.created = input, true
	return authrepo.RefreshSession{}, nil
}

func TestLoginCreatesHashedRefreshSession(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := &loginRepository{credentials: authrepo.UserCredentials{
		User: authrepo.User{ID: "user-1", Username: "alice"}, PasswordHash: string(hash),
	}}
	manager := token.NewJWTManager("secret", "issuer", "audience", time.Minute)
	service := authservice.NewAuthService(repo, nil, manager, nil, authservice.OAuthConfig{})
	before := time.Now()
	result, err := service.Login(context.Background(), authservice.LoginInput{Login: "alice", Password: "correct", UserAgent: "browser"})
	if err != nil {
		t.Fatal(err)
	}
	if repo.login != "alice" || repo.revokedUser != "user-1" || repo.revokedAgent == nil || *repo.revokedAgent != "browser" || !repo.created {
		t.Fatalf("browser session was not rotated: %+v", repo)
	}
	claims, err := manager.VerifyAccessToken(result.AccessToken)
	if err != nil || claims.UserID != "user-1" || claims.UserAgent != "browser" {
		t.Fatalf("invalid access token: %+v, err=%v", claims, err)
	}
	wantHash := sha256.Sum256([]byte(result.RefreshToken))
	if repo.session.TokenHash != hex.EncodeToString(wantHash[:]) || repo.session.TokenHash == result.RefreshToken {
		t.Fatal("refresh session must contain only the token hash")
	}
	if repo.session.UserID != "user-1" || repo.session.UserAgent == nil || *repo.session.UserAgent != "browser" {
		t.Fatalf("unexpected refresh session: %+v", repo.session)
	}
	if expiry := repo.session.ExpiresAt; expiry.Before(before.Add(30*24*time.Hour)) || expiry.After(time.Now().Add(30*24*time.Hour)) {
		t.Fatalf("unexpected refresh expiry: %v", expiry)
	}
	if result.User.ID != "user-1" || result.TokenType != "Bearer" || result.ExpiresIn != 15*60 || result.RefreshExpiresIn != 30*24*60*60 {
		t.Fatalf("unexpected login result: %+v", result)
	}
}

func TestLoginRejectsInvalidCredentials(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name      string
		lookupErr error
	}{
		{"unknown user", authrepo.ErrNotFound},
		{"wrong password", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &loginRepository{credentials: authrepo.UserCredentials{PasswordHash: string(hash)}, lookupErr: tc.lookupErr}
			service := authservice.NewAuthService(repo, nil, token.NewJWTManager("secret", "issuer", "audience", time.Minute), nil, authservice.OAuthConfig{})
			_, err := service.Login(context.Background(), authservice.LoginInput{Login: "alice", Password: "wrong"})
			if !errors.Is(err, authservice.ErrInvalidCredentials) || repo.created || repo.revokedUser != "" {
				t.Fatalf("expected no session on invalid credentials: err=%v, repo=%+v", err, repo)
			}
		})
	}
}
