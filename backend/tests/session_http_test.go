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

type sessionService struct {
	authservice.AuthService
	input         authservice.LogoutInput
	logoutCalled  bool
	refreshCalled bool
	logoutErr     error
	refreshErr    error
	refreshResult authservice.RefreshResult
}

func (s *sessionService) Logout(_ context.Context, input authservice.LogoutInput) error {
	s.input, s.logoutCalled = input, true
	return s.logoutErr
}

func (s *sessionService) Refresh(_ context.Context, input authservice.LogoutInput) (authservice.RefreshResult, error) {
	s.input, s.refreshCalled = input, true
	return s.refreshResult, s.refreshErr
}

func sessionRequest(t *testing.T, service *sessionService, method, path string, cookie bool) *httptest.ResponseRecorder {
	t.Helper()
	router := gin.New()
	handlerauth.NewHandler(service).RegisterRoutes(router.Group("/api/v1"), func(c *gin.Context) { c.Next() })
	request := httptest.NewRequest(method, path, nil)
	request.Header.Set("Authorization", "Bearer access-token")
	if cookie {
		// #nosec G124 -- This is an incoming test request cookie, not a Set-Cookie response.
		request.AddCookie(&http.Cookie{Name: "refresh_token", Value: "old-refresh"})
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestRefreshHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &sessionService{refreshResult: authservice.RefreshResult{
		AccessToken: "new-access", RefreshToken: "new-refresh", TokenType: "Bearer", ExpiresIn: 900, RefreshExpiresIn: 3600,
	}}
	response := sessionRequest(t, service, http.MethodGet, "/api/v1/auth/refresh", true)
	if response.Code != http.StatusOK || !service.refreshCalled || service.input.RefreshSession != "old-refresh" || service.input.AccessToken != "access-token" {
		t.Fatalf("unexpected refresh: status=%d, input=%+v", response.Code, service.input)
	}
	if response.Header().Get("Authorization") != "Bearer new-access" || !strings.Contains(response.Body.String(), `"accessToken":"new-access"`) || strings.Contains(response.Body.String(), "new-refresh") {
		t.Fatalf("unexpected refresh response: %s", response.Body.String())
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Value != "new-refresh" || !cookies[0].HttpOnly {
		t.Fatalf("unexpected refresh cookie: %+v", cookies)
	}
}

func TestRefreshHTTPRejectsMissingAndInvalidSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name   string
		cookie bool
		err    error
		code   string
		called bool
	}{
		{"missing", false, nil, "refresh_session_required", false},
		{"invalid", true, authservice.ErrInvalidRefreshSession, "invalid_refresh_session", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := &sessionService{refreshErr: tc.err}
			response := sessionRequest(t, service, http.MethodGet, "/api/v1/auth/refresh", tc.cookie)
			if response.Code != http.StatusUnauthorized || service.refreshCalled != tc.called || !strings.Contains(response.Body.String(), tc.code) {
				t.Fatalf("unexpected response: status=%d, called=%v, body=%s", response.Code, service.refreshCalled, response.Body.String())
			}
		})
	}
}

func TestLogoutHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name   string
		cookie bool
		err    error
		status int
		called bool
	}{
		{"success", true, nil, http.StatusNoContent, true},
		{"missing", false, nil, http.StatusUnauthorized, false},
		{"invalid", true, authservice.ErrInvalidRefreshSession, http.StatusUnauthorized, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := &sessionService{logoutErr: tc.err}
			response := sessionRequest(t, service, http.MethodDelete, "/api/v1/auth/logout", tc.cookie)
			if response.Code != tc.status || service.logoutCalled != tc.called {
				t.Fatalf("unexpected logout: status=%d, called=%v", response.Code, service.logoutCalled)
			}
			if tc.called && (service.input.RefreshSession != "old-refresh" || service.input.AccessToken != "access-token") {
				t.Fatalf("unexpected logout input: %+v", service.input)
			}
			if tc.cookie && len(response.Result().Cookies()) != 1 {
				t.Fatalf("expected refresh cookie to be cleared: %+v", response.Result().Cookies())
			}
		})
	}
}
