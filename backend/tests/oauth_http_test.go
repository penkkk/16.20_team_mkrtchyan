package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	handlerauth "opd/internal/handler/auth"
	authservice "opd/internal/service/auth"

	"github.com/gin-gonic/gin"
)

type oauthHTTPService struct {
	authservice.AuthService
	startInput     authservice.StartOAuthInput
	startResult    authservice.StartOAuthResult
	startErr       error
	callbackInput  authservice.CompleteOAuthInput
	callbackResult authservice.CompleteOAuthResult
	callbackErr    error
	completeInput  authservice.CompleteOAuthRegistrationInput
	completeResult authservice.LoginResult
	completeErr    error
}

func (s *oauthHTTPService) StartOAuth(_ context.Context, input authservice.StartOAuthInput) (authservice.StartOAuthResult, error) {
	s.startInput = input
	return s.startResult, s.startErr
}

func (s *oauthHTTPService) CompleteOAuth(_ context.Context, input authservice.CompleteOAuthInput) (authservice.CompleteOAuthResult, error) {
	s.callbackInput = input
	return s.callbackResult, s.callbackErr
}

func (s *oauthHTTPService) CompleteOAuthRegistration(_ context.Context, input authservice.CompleteOAuthRegistrationInput) (authservice.LoginResult, error) {
	s.completeInput = input
	return s.completeResult, s.completeErr
}

func oauthRequest(t *testing.T, service *oauthHTTPService, method, path, body, cookieName string) *httptest.ResponseRecorder {
	t.Helper()
	router := gin.New()
	handlerauth.NewHandler(service).RegisterRoutes(router.Group("/api/v1"), func(c *gin.Context) { c.Next() })
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "browser")
	if cookieName != "" {
		// #nosec G124 -- This is an incoming test request cookie, not a Set-Cookie response.
		request.AddCookie(&http.Cookie{Name: cookieName, Value: "saved-token"})
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestOAuthStartHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &oauthHTTPService{startResult: authservice.StartOAuthResult{RedirectURL: "https://accounts.google.com/auth", State: "state-1", MaxAge: 600}}
	response := oauthRequest(t, service, http.MethodGet, "/api/v1/auth/google/start?return_url=%2Fprofile", "", "")
	if response.Code != http.StatusFound || response.Header().Get("Location") != service.startResult.RedirectURL || service.startInput.Provider != "google" || service.startInput.ReturnURL != "/profile" {
		t.Fatalf("unexpected OAuth start: status=%d, input=%+v", response.Code, service.startInput)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "oauth_state" || cookies[0].Value != "state-1" || !cookies[0].HttpOnly {
		t.Fatalf("unexpected state cookie: %+v", cookies)
	}
	service.startErr = authservice.ErrUnsupportedOAuthProvider
	response = oauthRequest(t, service, http.MethodGet, "/api/v1/auth/unknown/start", "", "")
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "unsupported_oauth_provider") {
		t.Fatalf("unexpected provider error: %s", response.Body.String())
	}
}

func TestOAuthCallbackHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &oauthHTTPService{callbackResult: authservice.CompleteOAuthResult{LoginResult: authservice.LoginResult{
		AccessToken: "access", RefreshToken: "refresh", TokenType: "Bearer", RefreshExpiresIn: 3600, ReturnURL: "/profile",
	}}}
	path := "/api/v1/auth/google/callback?code=code-1&state=state-1"
	response := oauthRequest(t, service, http.MethodGet, path, "", "oauth_state")
	if response.Code != http.StatusFound || response.Header().Get("Location") != "/profile" || response.Header().Get("Authorization") != "Bearer access" ||
		service.callbackInput.Provider != "google" || service.callbackInput.Code != "code-1" || service.callbackInput.State != "state-1" || service.callbackInput.BrowserState != "saved-token" {
		t.Fatalf("unexpected callback: status=%d, input=%+v", response.Code, service.callbackInput)
	}
	if cookies := response.Result().Cookies(); len(cookies) != 2 || cookies[1].Name != "refresh_token" || cookies[1].Value != "refresh" {
		t.Fatalf("unexpected callback cookies: %+v", cookies)
	}
	pending := &oauthHTTPService{callbackResult: authservice.CompleteOAuthResult{PendingToken: "pending-1"}}
	response = oauthRequest(t, pending, http.MethodGet, path, "", "oauth_state")
	if response.Code != http.StatusFound || response.Header().Get("Location") != "/register?oauth=choose-username" {
		t.Fatalf("unexpected pending registration redirect: %d", response.Code)
	}
	if cookies := response.Result().Cookies(); len(cookies) != 2 || cookies[1].Name != "oauth_pending" || cookies[1].Value != "pending-1" {
		t.Fatalf("unexpected pending cookies: %+v", cookies)
	}
}

func TestOAuthCallbackRejectsInvalidState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	path := "/api/v1/auth/google/callback?code=code-1&state=state-1"
	response := oauthRequest(t, &oauthHTTPService{}, http.MethodGet, path, "", "")
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "oauth_state_missing") {
		t.Fatalf("missing state: %s", response.Body.String())
	}
	response = oauthRequest(t, &oauthHTTPService{}, http.MethodGet, path+"&error=access_denied", "", "oauth_state")
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "oauth_provider_error") {
		t.Fatalf("provider error: %s", response.Body.String())
	}
	response = oauthRequest(t, &oauthHTTPService{callbackErr: authservice.ErrOAuthStateInvalid}, http.MethodGet, path, "", "oauth_state")
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "invalid_oauth_state") {
		t.Fatalf("invalid state: %s", response.Body.String())
	}
}

func TestOAuthRegistrationHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &oauthHTTPService{completeResult: authservice.LoginResult{
		AccessToken: "access", RefreshToken: "refresh", TokenType: "Bearer", RefreshExpiresIn: 3600,
		User: authservice.User{ID: "user-1", Username: "alice"},
	}}
	path := "/api/v1/auth/oauth/complete"
	response := oauthRequest(t, service, http.MethodPost, path, `{"username":"alice"}`, "oauth_pending")
	if response.Code != http.StatusCreated || service.completeInput.PendingToken != "saved-token" || service.completeInput.Username != "alice" || service.completeInput.UserAgent != "browser" {
		t.Fatalf("unexpected OAuth registration: status=%d, input=%+v", response.Code, service.completeInput)
	}
	response = oauthRequest(t, &oauthHTTPService{}, http.MethodPost, path, `{"username":"alice"}`, "")
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "oauth_registration_not_found") {
		t.Fatalf("missing pending cookie: %s", response.Body.String())
	}
	response = oauthRequest(t, &oauthHTTPService{completeErr: authservice.NewFieldConflictError([]string{"username"})}, http.MethodPost, path, `{"username":"alice"}`, "oauth_pending")
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "username_already_exists") {
		t.Fatalf("conflicting username: %s", response.Body.String())
	}
}
