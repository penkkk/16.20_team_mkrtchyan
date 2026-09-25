package auth

import (
	"errors"
	"net/http"

	authservice "opd/internal/service/auth"

	"github.com/gin-gonic/gin"
)

const (
	refreshTokenCookieName = "refresh_token"
	refreshTokenCookiePath = "/api/v1/auth"
)

func (h *Handler) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.Login(c.Request.Context(), authservice.LoginInput{
		Login:    req.Login,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, authservice.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid login or password"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		refreshTokenCookieName,
		result.RefreshToken,
		result.RefreshExpiresIn,
		refreshTokenCookiePath,
		"",
		true,
		true,
	)

	setAccessTokenHeader(c, result)
	c.JSON(http.StatusOK, loginResponseFromService(result))
}

func (h *Handler) logout(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "logout is not implemented"})
}

func (h *Handler) register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	})
	if err != nil {
		var conflictErr *authservice.FieldConflictError
		switch {
		case errors.As(err, &conflictErr):
			c.JSON(http.StatusConflict, gin.H{
				"error":  "fields already taken",
				"fields": conflictFieldsResponse(conflictErr.Fields),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		refreshTokenCookieName,
		result.RefreshToken,
		result.RefreshExpiresIn,
		refreshTokenCookiePath,
		"",
		true,
		true,
	)

	setAccessTokenHeader(c, result)
	c.JSON(http.StatusCreated, loginResponseFromService(result))
}

func (h *Handler) refresh(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "refresh is not implemented"})
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

func setAccessTokenHeader(c *gin.Context, result authservice.LoginResult) {
	c.Header("Authorization", result.TokenType+" "+result.AccessToken)
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
