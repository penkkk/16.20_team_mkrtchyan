package tests

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"opd/internal/handler/middleware"
	authrepo "opd/internal/repository/auth"
	"opd/internal/service/token"

	"github.com/gin-gonic/gin"
)

type roleReaderStub struct {
	role   string
	err    error
	userID string
}

func (s *roleReaderStub) GetUserRoleByID(_ context.Context, input authrepo.GetRoleByUserID) (string, error) {
	s.userID = input.UserID
	return s.role, s.err
}

func TestRequireAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, userID, role string
		err                error
		want               int
	}{
		{name: "admin", userID: "user-id", role: "admin", want: http.StatusNoContent},
		{name: "user", userID: "user-id", role: "user", want: http.StatusForbidden},
		{name: "missing identity", want: http.StatusUnauthorized},
		{name: "deleted user", userID: "user-id", err: authrepo.ErrNotFound, want: http.StatusUnauthorized},
		{name: "database error", userID: "user-id", err: errors.New("database unavailable"), want: http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reader := &roleReaderStub{role: tc.role, err: tc.err}
			router := gin.New()
			router.GET("/admin", func(c *gin.Context) {
				if tc.userID != "" {
					c.Set(middleware.UserIDKey, tc.userID)
				}
				c.Next()
			}, middleware.RequireAdmin(reader), func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/admin", nil))
			if response.Code != tc.want || reader.userID != tc.userID {
				t.Fatalf("status=%d, queried user=%q; want status=%d, user=%q", response.Code, reader.userID, tc.want, tc.userID)
			}
		})
	}
}

func TestAuthMiddlewareRejectsMissingAndInvalidTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := token.NewJWTManager("secret", "issuer", "audience", time.Minute)
	for _, tc := range []struct{ name, header string }{
		{"missing", ""},
		{"wrong scheme", "Basic abc"},
		{"empty bearer", "Bearer "},
		{"invalid signature", "Bearer invalid-token"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/private", middleware.AuthMiddleware(manager, nil), func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})
			request := httptest.NewRequest(http.MethodGet, "/private", nil)
			request.Header.Set("Authorization", tc.header)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("unexpected status: %d", response.Code)
			}
		})
	}
}
