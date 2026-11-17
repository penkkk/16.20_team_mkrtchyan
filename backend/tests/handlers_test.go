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

type testLoginService struct {
	authservice.AuthService
	input  authservice.LoginInput
	called bool
	result authservice.LoginResult
	err    error
}

func (s *testLoginService) Login(_ context.Context, input authservice.LoginInput) (authservice.LoginResult, error) {
	s.input, s.called = input, true
	return s.result, s.err
}

func loginRequest(t *testing.T, service *testLoginService, body string) *httptest.ResponseRecorder {
	t.Helper()
	router := gin.New()
	handlerauth.NewHandler(service).RegisterRoutes(router.Group("/api/v1"), func(c *gin.Context) { c.Next() })
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "browser")
	request.Header.Set("X-Forwarded-Proto", "https")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestLoginHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &testLoginService{result: authservice.LoginResult{
		AccessToken: "access-token", RefreshToken: "refresh-secret", TokenType: "Bearer",
		ExpiresIn: 900, RefreshExpiresIn: 3600, User: authservice.User{ID: "user-1"},
	}}
	response := loginRequest(t, service, `{"login":"alice","password":"secret"}`)
	if response.Code != http.StatusOK || !service.called || service.input.Login != "alice" || service.input.Password != "secret" || service.input.UserAgent != "browser" {
		t.Fatalf("unexpected HTTP result: status=%d, input=%+v", response.Code, service.input)
	}
	if response.Header().Get("Authorization") != "Bearer access-token" {
		t.Fatalf("unexpected Authorization header: %q", response.Header().Get("Authorization"))
	}
	if body := response.Body.String(); !strings.Contains(body, `"accessToken":"access-token"`) || strings.Contains(body, "refresh-secret") {
		t.Fatalf("unexpected response body: %s", body)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "refresh_token" || cookies[0].Value != "refresh-secret" ||
		!cookies[0].HttpOnly || !cookies[0].Secure || cookies[0].Path != "/api/v1/auth" || cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatalf("unexpected refresh cookie: %+v", cookies)
	}
}

func TestLoginHTTPErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, body string
		err        error
		status     int
		called     bool
	}{
		{"missing password", `{"login":"alice"}`, nil, http.StatusBadRequest, false},
		{"bad credentials", `{"login":"alice","password":"wrong"}`, authservice.ErrInvalidCredentials, http.StatusUnauthorized, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := &testLoginService{err: tc.err}
			response := loginRequest(t, service, tc.body)
			if response.Code != tc.status || service.called != tc.called || len(response.Result().Cookies()) != 0 {
				t.Fatalf("status=%d, called=%v, cookies=%v", response.Code, service.called, response.Result().Cookies())
			}
		})
	}
}
