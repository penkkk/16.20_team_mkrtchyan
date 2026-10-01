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
	oauthStateCookieName   = "oauth_state"
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

func (h *Handler) startOAuth(c *gin.Context) {
	provider := c.Param("provider")
	returnURL := c.Query("return_url")

	result, err := h.service.StartOAuth(c.Request.Context(), authservice.StartOAuthInput{
		Provider:  provider,
		ReturnURL: returnURL,
	})
	if err != nil {
		switch {
		case errors.Is(err, authservice.ErrUnsupportedOAuthProvider):
			c.JSON(http.StatusBadRequest, gin.H{errorKey: "unsupported oauth provider"})
		case errors.Is(err, authservice.ErrOAuthClientIDRequired):
			c.JSON(http.StatusInternalServerError, gin.H{errorKey: "oauth client id is not configured"})
		case errors.Is(err, authservice.ErrOAuthRedirectURIRequired):
			c.JSON(http.StatusInternalServerError, gin.H{errorKey: "oauth redirect uri is not configured"})
		case errors.Is(err, authservice.ErrNotImplemented):
			c.JSON(http.StatusNotImplemented, gin.H{errorKey: "oauth is not implemented"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{errorKey: internalServerErrorMsg})
		}
		return
	}

	setOAuthStateCookie(c, result.State, result.MaxAge)
	c.Redirect(http.StatusFound, result.RedirectURL)
}

func (h *Handler) oauthCallback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")
	providerError := c.Query("error")
	if providerError != "" {
		clearOAuthStateCookie(c)
		c.JSON(http.StatusBadRequest, gin.H{
			errorKey:            "oauth provider error",
			"code":              providerError,
			"error_description": c.Query("error_description"),
		})
		return
	}

	browserState, err := c.Cookie(oauthStateCookieName)
	if err != nil {
		clearOAuthStateCookie(c)
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "oauth state is missing"})
		return
	}

	result, err := h.service.CompleteOAuth(c.Request.Context(), authservice.CompleteOAuthInput{
		Provider:     provider,
		Code:         code,
		State:        state,
		BrowserState: browserState,
		UserAgent:    c.Request.UserAgent(),
	})
	if err != nil {
		clearOAuthStateCookie(c)
		switch {
		case errors.Is(err, authservice.ErrUnsupportedOAuthProvider):
			c.JSON(http.StatusBadRequest, gin.H{errorKey: "unsupported oauth provider"})
		case errors.Is(err, authservice.ErrOAuthStateInvalid):
			c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid oauth state"})
		case errors.Is(err, authservice.ErrOAuthCodeRequired):
			c.JSON(http.StatusBadRequest, gin.H{errorKey: "oauth code is required"})
		case errors.Is(err, authservice.ErrOAuthProviderMismatch):
			c.JSON(http.StatusBadRequest, gin.H{errorKey: "oauth provider mismatch"})
		case errors.Is(err, authservice.ErrOAuthNonceMismatch):
			c.JSON(http.StatusBadRequest, gin.H{errorKey: "oauth nonce mismatch"})
		case errors.Is(err, authservice.ErrOAuthEmailAlreadyUsed):
			c.JSON(http.StatusConflict, gin.H{
				errorKey: "email already registered; link oauth provider in profile",
				"code":   "oauth_email_already_used",
			})
		case errors.Is(err, authservice.ErrOAuthUsernameRequired):
			c.JSON(http.StatusConflict, gin.H{
				errorKey: "username is required",
				"code":   "oauth_username_required",
			})
		case errors.Is(err, authservice.ErrNotImplemented):
			c.JSON(http.StatusNotImplemented, gin.H{errorKey: "oauth is not implemented"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{errorKey: internalServerErrorMsg})
		}
		return
	}

	clearOAuthStateCookie(c)
	setRefreshTokenCookie(c, result.RefreshToken, result.RefreshExpiresIn)
	setAccessTokenHeader(c, result.TokenType, result.AccessToken)
	if result.ReturnURL == "" {
		result.ReturnURL = "/"
	}
	c.Redirect(http.StatusFound, result.ReturnURL)
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
		secureCookie(c),
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
		secureCookie(c),
		true,
	)
}

func setOAuthStateCookie(c *gin.Context, state string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		oauthStateCookieName,
		state,
		maxAge,
		refreshTokenCookiePath,
		"",
		secureCookie(c),
		true,
	)
}

func clearOAuthStateCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		oauthStateCookieName,
		"",
		-1,
		refreshTokenCookiePath,
		"",
		secureCookie(c),
		true,
	)
}

func secureCookie(c *gin.Context) bool {
	return c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https")
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
