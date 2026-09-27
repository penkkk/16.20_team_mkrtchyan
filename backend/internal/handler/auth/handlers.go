package auth

import (
	"errors"
	"net/http"
	"strings"

	authservice "opd/internal/service/auth"

	"github.com/gin-gonic/gin"
)

const (
	refreshTokenCookieName = "refresh_token"
	refreshTokenCookiePath = "/api/v1/auth" //nolint:gosec
	errorKey               = "error"
	internalServerErrorMsg = "internal server error"
	bearerTokenPrefix      = "Bearer "
)

func (h *Handler) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: err.Error()})
		return
	}

	result, err := h.service.Login(c.Request.Context(), authservice.LoginInput{
		Login:     req.Login,
		Password:  req.Password,
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		switch {
		case errors.Is(err, authservice.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{errorKey: "invalid login or password"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{errorKey: internalServerErrorMsg})
		}
		return
	}

	setRefreshTokenCookie(c, result.RefreshToken, result.RefreshExpiresIn)
	setAccessTokenHeader(c, result.TokenType, result.AccessToken)
	c.JSON(http.StatusOK, loginResponseFromService(result))
}

func (h *Handler) logout(c *gin.Context) {
	refreshToken, err := c.Cookie(refreshTokenCookieName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{errorKey: "refresh session is required"})
		return
	}

	accessToken, _ := bearerTokenFromHeader(c)

	err = h.service.Logout(c.Request.Context(), authservice.LogoutInput{
		AccessToken:    accessToken,
		RefreshSession: refreshToken,
	})
	if err != nil {
		switch {
		case errors.Is(err, authservice.ErrInvalidRefreshSession):
			clearRefreshTokenCookie(c)
			c.JSON(http.StatusUnauthorized, gin.H{errorKey: "invalid refresh session"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{errorKey: "internal server error"})
		}
		return
	}

	clearRefreshTokenCookie(c)
	c.Status(http.StatusNoContent)
}

func (h *Handler) register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: err.Error()})
		return
	}

	tgUsername := optionalStringPtr(req.TgUsername)
	result, err := h.service.Register(c.Request.Context(), authservice.RegisterInput{
		Name:       req.Name,
		Password:   req.Password,
		Surname:    req.Surname,
		Email:      req.Email,
		TgUsername: tgUsername,
		Username:   req.Username,
		UserAgent:  c.Request.UserAgent(),
	})
	if err != nil {
		var conflictErr *authservice.FieldConflictError
		switch {
		case errors.As(err, &conflictErr):
			c.JSON(http.StatusConflict, gin.H{
				errorKey: "fields already taken",
				"fields": conflictFieldsResponse(conflictErr.Fields),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{errorKey: internalServerErrorMsg})
		}
		return
	}

	setRefreshTokenCookie(c, result.RefreshToken, result.RefreshExpiresIn)
	setAccessTokenHeader(c, result.TokenType, result.AccessToken)
	c.JSON(http.StatusCreated, loginResponseFromService(result))
}

func (h *Handler) refresh(c *gin.Context) {
	refreshToken, err := c.Cookie(refreshTokenCookieName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{errorKey: "refresh session is required"})
		return
	}

	accessToken, _ := bearerTokenFromHeader(c)

	result, err := h.service.Refresh(c.Request.Context(), authservice.LogoutInput{
		AccessToken:    accessToken,
		RefreshSession: refreshToken,
	})
	if err != nil {
		switch {
		case errors.Is(err, authservice.ErrInvalidRefreshSession):
			clearRefreshTokenCookie(c)
			c.JSON(http.StatusUnauthorized, gin.H{errorKey: "invalid refresh session"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{errorKey: "internal server error"})
		}
		return
	}

	setRefreshTokenCookie(c, result.RefreshToken, result.RefreshExpiresIn)
	setAccessTokenHeader(c, result.TokenType, result.AccessToken)
	c.JSON(http.StatusOK, refreshResponseFromService(result))
}

func loginResponseFromService(result authservice.LoginResult) LoginResponse {
	return LoginResponse{
		AccessToken:      result.AccessToken,
		TokenType:        result.TokenType,
		ExpiresIn:        result.ExpiresIn,
		RefreshExpiresIn: result.RefreshExpiresIn,
		User: UserResponse{
			ID:         result.User.ID,
			Username:   result.User.Username,
			Email:      result.User.Email,
			TgUsername: result.User.TgUsername,
			Name:       result.User.Name,
			Surname:    result.User.Surname,
		},
	}
}

func refreshResponseFromService(result authservice.RefreshResult) RefreshResponse {
	return RefreshResponse{
		AccessToken:      result.AccessToken,
		TokenType:        result.TokenType,
		ExpiresIn:        result.ExpiresIn,
		RefreshExpiresIn: result.RefreshExpiresIn,
	}
}

func setAccessTokenHeader(c *gin.Context, tokenType string, accessToken string) {
	c.Header("Authorization", tokenType+" "+accessToken)
}

func bearerTokenFromHeader(c *gin.Context) (string, bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", false
	}

	if !strings.HasPrefix(authHeader, bearerTokenPrefix) {
		return "", false
	}

	accessToken := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerTokenPrefix))
	if accessToken == "" {
		return "", false
	}

	return accessToken, true
}

func setRefreshTokenCookie(c *gin.Context, refreshToken string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		refreshTokenCookieName,
		refreshToken,
		maxAge,
		refreshTokenCookiePath,
		"",
		true,
		true,
	)
}

func clearRefreshTokenCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		refreshTokenCookieName,
		"",
		-1,
		refreshTokenCookiePath,
		"",
		true,
		true,
	)
}

func optionalStringPtr(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

func conflictFieldsResponse(fields []string) map[string]string {
	response := make(map[string]string, len(fields))
	for _, field := range fields {
		response[field] = "already_taken"
	}

	return response
}
