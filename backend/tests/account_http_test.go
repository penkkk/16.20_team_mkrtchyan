package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	handlerauth "opd/internal/handler/auth"
	"opd/internal/handler/middleware"
	authservice "opd/internal/service/auth"

	"github.com/gin-gonic/gin"
)

type accountService struct {
	authservice.AuthService
	registerInput  authservice.RegisterInput
	registerErr    error
	registerResult authservice.LoginResult
	registerCalled bool
	meInput        authservice.MeInput
	meResult       authservice.User
	meErr          error
	passwordInput  authservice.AddPasswordInput
	passwordErr    error
}

func (s *accountService) Register(_ context.Context, input authservice.RegisterInput) (authservice.LoginResult, error) {
	s.registerInput, s.registerCalled = input, true
	return s.registerResult, s.registerErr
}

func (s *accountService) Me(_ context.Context, input authservice.MeInput) (authservice.User, error) {
	s.meInput = input
	return s.meResult, s.meErr
}

func (s *accountService) AddPassword(_ context.Context, input authservice.AddPasswordInput) error {
	s.passwordInput = input
	return s.passwordErr
}

func accountRequest(t *testing.T, service *accountService, userID, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	router := gin.New()
	authRequired := func(c *gin.Context) {
		if userID != "" {
			c.Set(middleware.UserIDKey, userID)
		}
		c.Next()
	}
	handlerauth.NewHandler(service).RegisterRoutes(router.Group("/api/v1"), authRequired)
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "browser")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestRegisterHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const body = `{"username":"alice","email":"alice@example.com","name":"Alice","surname":"Doe","password":"secret"}`
	service := &accountService{registerResult: authservice.LoginResult{
		AccessToken: "access", RefreshToken: "refresh", TokenType: "Bearer", RefreshExpiresIn: 3600,
		User: authservice.User{ID: "user-1", Username: "alice"},
	}}
	response := accountRequest(t, service, "", http.MethodPost, "/api/v1/auth/register", body)
	if response.Code != http.StatusCreated || !service.registerCalled || service.registerInput.Username != "alice" || service.registerInput.Email != "alice@example.com" || service.registerInput.UserAgent != "browser" || service.registerInput.TgUsername != nil {
		t.Fatalf("unexpected registration: status=%d, input=%+v", response.Code, service.registerInput)
	}
	if response.Header().Get("Authorization") != "Bearer access" || strings.Contains(response.Body.String(), `"refreshToken"`) || len(response.Result().Cookies()) != 1 {
		t.Fatalf("unexpected registration response: %s", response.Body.String())
	}
	conflict := &accountService{registerErr: authservice.NewFieldConflictError([]string{"username"})}
	response = accountRequest(t, conflict, "", http.MethodPost, "/api/v1/auth/register", body)
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "username_already_exists") {
		t.Fatalf("unexpected conflict response: %s", response.Body.String())
	}
	invalid := &accountService{}
	response = accountRequest(t, invalid, "", http.MethodPost, "/api/v1/auth/register", `{"username":"alice"}`)
	if response.Code != http.StatusBadRequest || invalid.registerCalled {
		t.Fatalf("invalid registration reached service: %d", response.Code)
	}
}

func TestMeHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &accountService{meResult: authservice.User{ID: "user-1", Username: "alice"}}
	response := accountRequest(t, service, "user-1", http.MethodGet, "/api/v1/me", "")
	if response.Code != http.StatusOK || service.meInput.UserID != "user-1" || !strings.Contains(response.Body.String(), `"username":"alice"`) {
		t.Fatalf("unexpected profile response: status=%d, body=%s", response.Code, response.Body.String())
	}
	response = accountRequest(t, &accountService{}, "", http.MethodGet, "/api/v1/me", "")
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("missing user ID: status=%d", response.Code)
	}
	response = accountRequest(t, &accountService{meErr: authservice.ErrUserNotFound}, "user-1", http.MethodGet, "/api/v1/me", "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("missing user: status=%d", response.Code)
	}
}

func TestAddPasswordHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &accountService{}
	response := accountRequest(t, service, "user-1", http.MethodPost, "/api/v1/auth/password", `{"password":"new-secret"}`)
	if response.Code != http.StatusNoContent || service.passwordInput.UserID != "user-1" || service.passwordInput.Password != "new-secret" {
		t.Fatalf("unexpected add-password result: status=%d, input=%+v", response.Code, service.passwordInput)
	}
	response = accountRequest(t, &accountService{passwordErr: authservice.ErrPasswordAlreadySet}, "user-1", http.MethodPost, "/api/v1/auth/password", `{"password":"new-secret"}`)
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "password_already_set") {
		t.Fatalf("unexpected duplicate-password response: %s", response.Body.String())
	}
	response = accountRequest(t, &accountService{}, "", http.MethodPost, "/api/v1/auth/password", `{"password":"new-secret"}`)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("missing user ID: status=%d", response.Code)
	}
}
