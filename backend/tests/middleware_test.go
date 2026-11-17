package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"opd/internal/handler/middleware"
	"opd/internal/service/token"

	"github.com/gin-gonic/gin"
)

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
