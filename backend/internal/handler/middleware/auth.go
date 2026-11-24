package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"opd/internal/handler/apierror"
	authrepo "opd/internal/repository/auth"
	"opd/internal/service/token"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	unauthorizedMsg      = "unauthorized"
	UserIDKey            = "userID"
	redisBlackListPrefix = "auth:blacklist:"
)

func AuthMiddleware(tokens token.Manager, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			apierror.Abort(c, http.StatusUnauthorized, "unauthorized", unauthorizedMsg)
			return
		}
		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			apierror.Abort(c, http.StatusUnauthorized, "unauthorized", unauthorizedMsg)
			return
		}

		accessToken := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
		if accessToken == "" {
			apierror.Abort(c, http.StatusUnauthorized, "unauthorized", unauthorizedMsg)
			return
		}

		claims, err := tokens.VerifyAccessToken(accessToken)
		if err != nil {
			apierror.Abort(c, http.StatusUnauthorized, "unauthorized", unauthorizedMsg)
			return
		}

		key := redisBlackListPrefix + hashToken(accessToken)
		exists, err := redisClient.Exists(c.Request.Context(), key).Result()
		if err != nil {
			apierror.Abort(c, http.StatusInternalServerError, "internal_server_error", "internal server error")
			return
		}
		if exists > 0 {
			apierror.Abort(c, http.StatusUnauthorized, "unauthorized", unauthorizedMsg)
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Next()
	}
}

type RoleReader interface {
	GetUserRoleByID(ctx context.Context, input authrepo.GetRoleByUserID) (string, error)
}

// RequireAdmin must run after AuthMiddleware.
func RequireAdmin(roles RoleReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString(UserIDKey)
		if userID == "" {
			apierror.Abort(c, http.StatusUnauthorized, "unauthorized", unauthorizedMsg)
			return
		}

		role, err := roles.GetUserRoleByID(c.Request.Context(), authrepo.GetRoleByUserID{UserID: userID})
		if errors.Is(err, authrepo.ErrNotFound) {
			apierror.Abort(c, http.StatusUnauthorized, "unauthorized", unauthorizedMsg)
			return
		}
		if err != nil {
			apierror.Abort(c, http.StatusInternalServerError, "internal_server_error", "internal server error")
			return
		}
		if role != "admin" {
			apierror.Abort(c, http.StatusForbidden, "forbidden", "forbidden")
			return
		}
		c.Next()
	}
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
