//go:build integration

package tests

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	authhandler "opd/internal/handler/auth"
	"opd/internal/handler/middleware"
	"opd/internal/repository"
	authservice "opd/internal/service/auth"
	"opd/internal/service/token"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

func TestIntegrationAuthLifecycle(t *testing.T) {
	databaseURL, redisURL := os.Getenv("OPD_TEST_DATABASE_URL"), os.Getenv("OPD_TEST_REDIS_URL")
	if databaseURL == "" || redisURL == "" {
		t.Fatal("run tests/run-integration.sh to start isolated PostgreSQL and Redis")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	redisOptions, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Fatal(err)
	}
	client := redis.NewClient(redisOptions)
	defer client.Close()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatal(err)
	}
	migration, err := os.ReadFile("../migrations/000001_create_users_and_refresh_sessions.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, string(migration)); err != nil {
		t.Fatalf("apply schema: %v", err)
	}

	manager := token.NewJWTManager("integration-secret", "opd-test", "opd-test", 15*time.Minute)
	repos := repository.NewRepositoriesWithTxManager(pool)
	service := authservice.NewAuthService(repos.Auth, repos.TxManager, manager, client, authservice.OAuthConfig{})
	username := fmt.Sprintf("test_%d", time.Now().UnixNano())
	registered, err := service.Register(ctx, authservice.RegisterInput{
		Username: username, Email: username + "@example.com", Password: "secret-password", Name: "Test", Surname: "User", UserAgent: "browser",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := pool.Exec(cleanupCtx, "DELETE FROM auth.users WHERE id = $1", registered.User.ID); err != nil {
			t.Errorf("delete test user: %v", err)
		}
	}()
	var passwordHash string
	if err := pool.QueryRow(ctx, "SELECT password_hash FROM auth.user_passwords WHERE user_id = $1", registered.User.ID).Scan(&passwordHash); err != nil {
		t.Fatalf("read stored password: %v", err)
	}
	if passwordHash == "secret-password" || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("secret-password")) != nil {
		t.Fatal("database did not store a valid bcrypt password hash")
	}
	if _, err := service.Login(ctx, authservice.LoginInput{Login: username, Password: "wrong"}); !errors.Is(err, authservice.ErrInvalidCredentials) {
		t.Fatalf("wrong password was accepted: %v", err)
	}
	loggedIn, err := service.Login(ctx, authservice.LoginInput{Login: username, Password: "secret-password", UserAgent: "browser"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	var oldRevoked bool
	if err := pool.QueryRow(ctx, "SELECT revoked_at IS NOT NULL FROM auth.refresh_sessions WHERE token_hash = $1", refreshHash(registered.RefreshToken)).Scan(&oldRevoked); err != nil || !oldRevoked {
		t.Fatalf("login did not revoke the previous browser session: revoked=%v, err=%v", oldRevoked, err)
	}
	refreshed, err := service.Refresh(ctx, authservice.LogoutInput{RefreshSession: loggedIn.RefreshToken})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if refreshed.RefreshToken == loggedIn.RefreshToken || refreshed.AccessToken == "" {
		t.Fatal("refresh did not issue a new token pair")
	}
	if _, err := service.Refresh(ctx, authservice.LogoutInput{RefreshSession: loggedIn.RefreshToken}); !errors.Is(err, authservice.ErrInvalidRefreshSession) {
		t.Fatalf("revoked refresh token was reused: %v", err)
	}
	var active bool
	if err := pool.QueryRow(ctx, "SELECT revoked_at IS NULL FROM auth.refresh_sessions WHERE token_hash = $1", refreshHash(refreshed.RefreshToken)).Scan(&active); err != nil || !active {
		t.Fatalf("new refresh session is not active: active=%v, err=%v", active, err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	authhandler.NewHandler(service).RegisterRoutes(router.Group("/api/v1"), middleware.AuthMiddleware(manager, client))
	checkAccess := func() int {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
		request.Header.Set("Authorization", "Bearer "+refreshed.AccessToken)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		return response.Code
	}
	if status := checkAccess(); status != http.StatusOK {
		t.Fatalf("valid access token was rejected: %d", status)
	}
	if err := service.Logout(ctx, authservice.LogoutInput{RefreshSession: refreshed.RefreshToken, AccessToken: refreshed.AccessToken}); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if status := checkAccess(); status != http.StatusUnauthorized {
		t.Fatalf("logged-out access token remained valid: %d", status)
	}
	if err := service.Logout(ctx, authservice.LogoutInput{RefreshSession: refreshed.RefreshToken}); !errors.Is(err, authservice.ErrInvalidRefreshSession) {
		t.Fatalf("logged-out refresh token was reused: %v", err)
	}
	_ = client.Del(ctx, "auth:blacklist:"+refreshHash(refreshed.AccessToken)).Err()
}

func refreshHash(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}
