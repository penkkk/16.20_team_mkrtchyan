package auth

import (
	"errors"
	"net/http"
	"strings"

	"opd/internal/handler/apierror"
	"opd/internal/handler/middleware"
	authservice "opd/internal/service/auth"

	"github.com/gin-gonic/gin"
)

const (
	refreshTokenCookieName = "refresh_token"
	authCookiePath         = "/api/v1/auth" //nolint:gosec
	oauthStateCookieName   = "oauth_state"
	oauthPendingCookieName = "oauth_pending"
	internalServerErrorMsg = "internal server error"
	bearerTokenPrefix      = "Bearer "
)

// login godoc
// @Summary Login
// @Description Authenticates a user and returns access and refresh token metadata. Refresh token is also set as an HTTP-only cookie.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} apierror.Response
// @Failure 401 {object} apierror.Response
// @Failure 500 {object} apierror.Response
// @Router /auth/login [post]
func (h *Handler) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err)
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
			apierror.Respond(c, http.StatusUnauthorized, "invalid_credentials", "Invalid login or password.")
		default:
			respondInternalServerError(c)
		}
		return
	}

	setAuthCookie(c, refreshTokenCookieName, result.RefreshToken, result.RefreshExpiresIn)
	setAccessTokenHeader(c, result.TokenType, result.AccessToken)
	c.JSON(http.StatusOK, loginResponseFromService(result))
}

// logout godoc
// @Summary Logout
// @Description Invalidates the current refresh session and blacklists the access token when it is provided.
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Param Cookie header string true "refresh_token cookie"
// @Success 204
// @Failure 401 {object} apierror.Response
// @Failure 500 {object} apierror.Response
// @Router /auth/logout [delete]
func (h *Handler) logout(c *gin.Context) {
	refreshToken, err := c.Cookie(refreshTokenCookieName)
	if err != nil {
		apierror.Respond(c, http.StatusUnauthorized, "refresh_session_required", "Refresh session is required.")
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
			clearAuthCookie(c, refreshTokenCookieName)
			apierror.Respond(c, http.StatusUnauthorized, "invalid_refresh_session", "Invalid refresh session.")
		default:
			respondInternalServerError(c)
		}
		return
	}

	clearAuthCookie(c, refreshTokenCookieName)
	c.Status(http.StatusNoContent)
}

// register godoc
// @Summary Register
// @Description Creates a user and returns access and refresh token metadata. Refresh token is also set as an HTTP-only cookie.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration data"
// @Success 201 {object} LoginResponse
// @Failure 400 {object} apierror.Response
// @Failure 409 {object} apierror.Response
// @Failure 500 {object} apierror.Response
// @Router /auth/register [post]
func (h *Handler) register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err)
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
			code, message, details := fieldConflictError(conflictErr.Fields)
			apierror.Respond(c, http.StatusConflict, code, message, details...)
		default:
			respondInternalServerError(c)
		}
		return
	}

	setAuthCookie(c, refreshTokenCookieName, result.RefreshToken, result.RefreshExpiresIn)
	setAccessTokenHeader(c, result.TokenType, result.AccessToken)
	c.JSON(http.StatusCreated, loginResponseFromService(result))
}

// refresh godoc
// @Summary Refresh tokens
// @Description Issues a new access token using the refresh token cookie.
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Param Cookie header string true "refresh_token cookie"
// @Success 200 {object} RefreshResponse
// @Failure 401 {object} apierror.Response
// @Failure 500 {object} apierror.Response
// @Router /auth/refresh [get]
func (h *Handler) refresh(c *gin.Context) {
	refreshToken, err := c.Cookie(refreshTokenCookieName)
	if err != nil {
		apierror.Respond(c, http.StatusUnauthorized, "refresh_session_required", "Refresh session is required.")
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
			clearAuthCookie(c, refreshTokenCookieName)
			apierror.Respond(c, http.StatusUnauthorized, "invalid_refresh_session", "Invalid refresh session.")
		default:
			respondInternalServerError(c)
		}
		return
	}

	setAuthCookie(c, refreshTokenCookieName, result.RefreshToken, result.RefreshExpiresIn)
	setAccessTokenHeader(c, result.TokenType, result.AccessToken)
	c.JSON(http.StatusOK, refreshResponseFromService(result))
}

// startOAuth godoc
// @Summary Start OAuth
// @Description Starts an OAuth flow for the selected provider and redirects to the provider authorization page.
// @Tags auth
// @Produce json
// @Param provider path string true "OAuth provider" Enums(google,yandex)
// @Param return_url query string false "URL to redirect to after successful OAuth callback"
// @Success 302
// @Failure 400 {object} apierror.Response
// @Failure 500 {object} apierror.Response
// @Failure 501 {object} apierror.Response
// @Router /auth/{provider}/start [get]
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
			apierror.Respond(c, http.StatusBadRequest, "unsupported_oauth_provider", "Unsupported OAuth provider.")
		case errors.Is(err, authservice.ErrOAuthClientIDRequired):
			apierror.Respond(c, http.StatusInternalServerError, "oauth_client_id_not_configured", "OAuth client ID is not configured.")
		case errors.Is(err, authservice.ErrOAuthRedirectURIRequired):
			apierror.Respond(c, http.StatusInternalServerError, "oauth_redirect_uri_not_configured", "OAuth redirect URI is not configured.")
		case errors.Is(err, authservice.ErrNotImplemented):
			apierror.Respond(c, http.StatusNotImplemented, "oauth_not_implemented", "OAuth is not implemented.")
		default:
			respondInternalServerError(c)
		}
		return
	}

	setAuthCookie(c, oauthStateCookieName, result.State, result.MaxAge)
	c.Redirect(http.StatusFound, result.RedirectURL)
}

// oauthCallback godoc
// @Summary Complete OAuth
// @Description Completes an OAuth flow, sets session cookies when authentication succeeds, and redirects to the return URL.
// @Tags auth
// @Produce json
// @Param provider path string true "OAuth provider" Enums(google,yandex)
// @Param code query string false "OAuth authorization code"
// @Param state query string false "OAuth state"
// @Param error query string false "OAuth provider error code"
// @Param error_description query string false "OAuth provider error description"
// @Param Cookie header string true "oauth_state cookie"
// @Success 302
// @Failure 400 {object} apierror.Response
// @Failure 409 {object} apierror.Response
// @Failure 500 {object} apierror.Response
// @Failure 501 {object} apierror.Response
// @Router /auth/{provider}/callback [get]
func (h *Handler) oauthCallback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")
	providerError := c.Query("error")
	if providerError != "" {
		clearAuthCookie(c, oauthStateCookieName)
		apierror.Respond(
			c,
			http.StatusBadRequest,
			"oauth_provider_error",
			"OAuth provider returned an error.",
			apierror.Detail{Field: "error", Message: providerError},
			apierror.Detail{Field: "error_description", Message: c.Query("error_description")},
		)
		return
	}

	browserState, err := c.Cookie(oauthStateCookieName)
	if err != nil {
		clearAuthCookie(c, oauthStateCookieName)
		apierror.Respond(c, http.StatusBadRequest, "oauth_state_missing", "OAuth state is missing.")
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
		clearAuthCookie(c, oauthStateCookieName)
		respondOAuthCallbackError(c, err)
		return
	}

	clearAuthCookie(c, oauthStateCookieName)

	if result.PendingToken != "" {
		setAuthCookie(c, oauthPendingCookieName, result.PendingToken, int(authservice.OAuthAttemptTTL.Seconds()))
		c.Redirect(http.StatusFound, "/register?oauth=choose-username")
		return
	}

	if result.RefreshToken != "" {
		setAuthCookie(c, refreshTokenCookieName, result.RefreshToken, result.RefreshExpiresIn)
	}
	if result.AccessToken != "" {
		setAccessTokenHeader(c, result.TokenType, result.AccessToken)
	}
	if result.ReturnURL == "" {
		result.ReturnURL = "/"
	}
	c.Redirect(http.StatusFound, result.ReturnURL)
}

// completeOAuthRegistration godoc
// @Summary Complete OAuth registration
// @Description Creates a user from a pending OAuth registration after the user chooses a username.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body CompleteOAuthRegistrationRequest true "Username"
// @Param Cookie header string true "oauth_pending cookie"
// @Success 201 {object} LoginResponse
// @Failure 400 {object} apierror.Response
// @Failure 409 {object} apierror.Response
// @Failure 500 {object} apierror.Response
// @Router /auth/oauth/complete [post]
func (h *Handler) completeOAuthRegistration(c *gin.Context) {
	pendingToken, err := c.Cookie(oauthPendingCookieName)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "oauth_registration_not_found", "OAuth registration has expired. Start again.")
		return
	}

	var req CompleteOAuthRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err)
		return
	}

	result, err := h.service.CompleteOAuthRegistration(c.Request.Context(), authservice.CompleteOAuthRegistrationInput{
		PendingToken: pendingToken,
		Username:     req.Username,
		UserAgent:    c.Request.UserAgent(),
	})
	if err != nil {
		var conflictErr *authservice.FieldConflictError
		switch {
		case errors.Is(err, authservice.ErrOAuthPendingRegistrationNotFound):
			clearAuthCookie(c, oauthPendingCookieName)
			apierror.Respond(c, http.StatusBadRequest, "oauth_registration_not_found", "OAuth registration has expired. Start again.")
		case errors.As(err, &conflictErr):
			code, message, details := fieldConflictError(conflictErr.Fields)
			apierror.Respond(c, http.StatusConflict, code, message, details...)
		case errors.Is(err, authservice.ErrOAuthIdentityAlreadyLinked):
			clearAuthCookie(c, oauthPendingCookieName)
			apierror.Respond(c, http.StatusConflict, "oauth_identity_already_linked", "OAuth identity is already linked.")
		default:
			respondInternalServerError(c)
		}
		return
	}

	clearAuthCookie(c, oauthPendingCookieName)
	setAuthCookie(c, refreshTokenCookieName, result.RefreshToken, result.RefreshExpiresIn)
	setAccessTokenHeader(c, result.TokenType, result.AccessToken)
	c.JSON(http.StatusCreated, loginResponseFromService(result))
}

func respondOAuthCallbackError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, authservice.ErrUnsupportedOAuthProvider):
		apierror.Respond(c, http.StatusBadRequest, "unsupported_oauth_provider", "Unsupported OAuth provider.")
	case errors.Is(err, authservice.ErrOAuthStateInvalid):
		apierror.Respond(c, http.StatusBadRequest, "invalid_oauth_state", "Invalid OAuth state.")
	case errors.Is(err, authservice.ErrOAuthCodeRequired):
		apierror.Respond(c, http.StatusBadRequest, "oauth_code_required", "OAuth code is required.")
	case errors.Is(err, authservice.ErrOAuthProviderMismatch):
		apierror.Respond(c, http.StatusBadRequest, "oauth_provider_mismatch", "OAuth provider mismatch.")
	case errors.Is(err, authservice.ErrOAuthNonceMismatch):
		apierror.Respond(c, http.StatusBadRequest, "oauth_nonce_mismatch", "OAuth nonce mismatch.")
	case errors.Is(err, authservice.ErrOAuthEmailAlreadyUsed):
		apierror.Respond(c, http.StatusConflict, "oauth_email_already_used", "Email is already registered; link OAuth provider in profile.")
	case errors.Is(err, authservice.ErrOAuthUsernameRequired):
		apierror.Respond(c, http.StatusConflict, "oauth_username_required", "Username is required.")
	case errors.Is(err, authservice.ErrOAuthIdentityAlreadyLinked):
		apierror.Respond(c, http.StatusConflict, "oauth_identity_already_linked", "OAuth identity is already linked.")
	case errors.Is(err, authservice.ErrOAuthProviderAlreadyLinked):
		apierror.Respond(c, http.StatusConflict, "oauth_provider_already_linked", "OAuth provider is already linked.")
	case errors.Is(err, authservice.ErrNotImplemented):
		apierror.Respond(c, http.StatusNotImplemented, "oauth_not_implemented", "OAuth is not implemented.")
	default:
		respondInternalServerError(c)
	}
}

// linkExternal godoc
// @Summary Start OAuth linking
// @Description Creates an OAuth authorization URL for linking an external provider to the authenticated user.
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Param provider path string true "OAuth provider" Enums(google,yandex)
// @Param return_url query string false "URL to redirect to after successful OAuth callback"
// @Success 200 {object} ExternalLinkResponse
// @Failure 400 {object} apierror.Response
// @Failure 401 {object} apierror.Response
// @Failure 500 {object} apierror.Response
// @Router /auth/{provider}/link/start [get]
func (h *Handler) linkExternal(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	if userID == "" {
		apierror.Respond(c, http.StatusUnauthorized, "unauthorized", "Unauthorized.")
		return
	}

	provider := c.Param("provider")
	returnURL := c.Query("return_url")

	result, err := h.service.StartOAuthLink(c.Request.Context(), authservice.StartOAuthLinkInput{
		Provider:  provider,
		ReturnURL: returnURL,
		UserID:    userID,
	})
	if err != nil {
		switch {
		case errors.Is(err, authservice.ErrUnsupportedOAuthProvider):
			apierror.Respond(c, http.StatusBadRequest, "unsupported_oauth_provider", "Unsupported OAuth provider.")
		case errors.Is(err, authservice.ErrOAuthClientIDRequired):
			apierror.Respond(c, http.StatusInternalServerError, "oauth_client_id_not_configured", "OAuth client ID is not configured.")
		case errors.Is(err, authservice.ErrOAuthRedirectURIRequired):
			apierror.Respond(c, http.StatusInternalServerError, "oauth_redirect_uri_not_configured", "OAuth redirect URI is not configured.")
		default:
			respondInternalServerError(c)
		}
		return
	}

	setAuthCookie(c, oauthStateCookieName, result.State, result.MaxAge)
	c.JSON(http.StatusOK, ExternalLinkResponse{
		RedirectURL: result.RedirectURL,
	})
}

// addPassword godoc
// @Summary Add password
// @Description Adds a password to the authenticated user when password authentication is not configured yet.
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AddPasswordRequest true "Password data"
// @Success 204
// @Failure 400 {object} apierror.Response
// @Failure 401 {object} apierror.Response
// @Failure 409 {object} apierror.Response
// @Failure 500 {object} apierror.Response
// @Router /auth/password [post]
func (h *Handler) addPassword(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	if userID == "" {
		apierror.Respond(c, http.StatusUnauthorized, "unauthorized", "Unauthorized.")
		return
	}

	var req AddPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err)
		return
	}

	err := h.service.AddPassword(c.Request.Context(), authservice.AddPasswordInput{
		UserID:   userID,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, authservice.ErrPasswordAlreadySet):
			apierror.Respond(c, http.StatusConflict, "password_already_set", "Password is already set.")
		default:
			respondInternalServerError(c)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// me godoc
// @Summary Current user
// @Description Returns the authenticated user's profile.
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserResponse
// @Failure 401 {object} apierror.Response
// @Failure 404 {object} apierror.Response
// @Failure 500 {object} apierror.Response
// @Router /me [get]
func (h *Handler) me(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	if userID == "" {
		apierror.Respond(c, http.StatusUnauthorized, "unauthorized", "Unauthorized.")
		return
	}

	user, err := h.service.Me(c.Request.Context(), authservice.MeInput{
		UserID: userID,
	})
	if err != nil {
		switch {
		case errors.Is(err, authservice.ErrUserNotFound):
			apierror.Respond(c, http.StatusNotFound, "user_not_found", "User not found.")
		default:
			respondInternalServerError(c)
		}
		return
	}

	c.JSON(http.StatusOK, userResponseFromService(user))
}

func loginResponseFromService(result authservice.LoginResult) LoginResponse {
	return LoginResponse{
		AccessToken:      result.AccessToken,
		TokenType:        result.TokenType,
		ExpiresIn:        result.ExpiresIn,
		RefreshExpiresIn: result.RefreshExpiresIn,
		User:             userResponseFromService(result.User),
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

func userResponseFromService(user authservice.User) UserResponse {
	return UserResponse{
		ID:         user.ID,
		Username:   user.Username,
		Email:      user.Email,
		TgUsername: user.TgUsername,
		Name:       user.Name,
		Surname:    user.Surname,
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

func clearAuthCookie(c *gin.Context, cookieName string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		cookieName,
		"",
		-1,
		authCookiePath,
		"",
		secureCookie(c),
		true,
	)
}

func setAuthCookie(c *gin.Context, cookieName string, value string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		cookieName,
		value,
		maxAge,
		authCookiePath,
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

func respondValidationError(c *gin.Context, err error) {
	apierror.Respond(
		c,
		http.StatusBadRequest,
		"invalid_request",
		"Request is invalid.",
		apierror.Detail{Message: err.Error()},
	)
}

func respondInternalServerError(c *gin.Context) {
	apierror.Respond(c, http.StatusInternalServerError, "internal_server_error", internalServerErrorMsg)
}

func fieldConflictError(fields []string) (string, string, []apierror.Detail) {
	if len(fields) == 1 {
		field := fields[0]
		return fieldConflictCode(field), fieldConflictMessage(field), []apierror.Detail{
			{
				Field:   field,
				Message: fieldConflictMessage(field),
			},
		}
	}

	details := make([]apierror.Detail, 0, len(fields))
	for _, field := range fields {
		details = append(details, apierror.Detail{
			Field:   field,
			Message: fieldConflictMessage(field),
		})
	}

	return "fields_already_taken", "One or more fields are already taken.", details
}

func fieldConflictCode(field string) string {
	switch field {
	case "username":
		return "username_already_exists"
	case "email":
		return "email_already_exists"
	case "tg_username":
		return "telegram_username_already_exists"
	default:
		return "field_already_exists"
	}
}

func fieldConflictMessage(field string) string {
	switch field {
	case "username":
		return "User with this username already exists."
	case "email":
		return "User with this email already exists."
	case "tg_username":
		return "User with this Telegram username already exists."
	default:
		return "Field value already exists."
	}
}
