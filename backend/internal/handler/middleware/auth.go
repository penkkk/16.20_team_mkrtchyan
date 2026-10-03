package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"opd/internal/handler/apierror"
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

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
