package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"

	"opd/internal/repository"
	authrepo "opd/internal/repository/auth"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

const (
	OAuthProviderGoogle = "google"
	OAuthProviderYandex = "yandex"
	OAuthModeLogin      = "login"
	OAuthModeLink       = "link"
	OAuthRedisKey       = "auth:oauth:"
	OAuthAttemptTTL     = 10 * time.Minute
)

type OAuthProvider interface {
	Name() string
	Scopes() []string
	AuthCodeURL(input OAuthAuthCodeURLInput) (string, error)
	ExchangeCode(ctx context.Context, input OAuthExchangeCodeInput) (OAuthTokens, error)
	FetchUser(ctx context.Context, tokens OAuthTokens, expectedNonce string) (OAuthProfile, error)
}

type OAuthProviderRegistry struct {
	providers map[string]OAuthProvider
}

func NewOAuthProviderRegistry(providers ...OAuthProvider) *OAuthProviderRegistry {
	registry := &OAuthProviderRegistry{
		providers: make(map[string]OAuthProvider, len(providers)),
	}

	for _, provider := range providers {
		registry.providers[provider.Name()] = provider
	}

	return registry
}

func (r *OAuthProviderRegistry) Get(name string) (OAuthProvider, bool) {
	provider, ok := r.providers[name]
	return provider, ok
}

type OAuthAuthCodeURLInput struct {
	RedirectURI   string
	State         string
	CodeChallenge string
	Nonce         string
	Scopes        []string
}

type OAuthExchangeCodeInput struct {
	Code         string
	CodeVerifier string
	RedirectURI  string
}

type OAuthTokens struct {
	AccessToken  string
	RefreshToken string
	IDToken      string
	TokenType    string
	ExpiresIn    int
}

type OAuthProfile struct {
	Provider       string
	ProviderUserID string
	Email          string
	FirstName      string
	LastName       string
	UsernameHint   string
	EmailVerified  bool
}

type OAuthAttempt struct {
	Provider     string `json:"provider"`
	CodeVerifier string `json:"code_verifier"`
	Nonce        string `json:"nonce,omitempty"`
	ReturnURL    string `json:"return_url,omitempty"`
	Mode         string `json:"mode"`
	UserID       string `json:"user_id,omitempty"`
}

func (s *authService) StartOAuth(ctx context.Context, input StartOAuthInput) (StartOAuthResult, error) {
	return s.startOAuthAttempt(ctx, input.Provider, input.ReturnURL, OAuthModeLogin, "")
}

func (s *authService) StartOAuthLink(ctx context.Context, input StartOAuthLinkInput) (StartOAuthResult, error) {
	if input.UserID == "" {
		return StartOAuthResult{}, ErrInvalidCredentials
	}
	return s.startOAuthAttempt(ctx, input.Provider, input.ReturnURL, OAuthModeLink, input.UserID)
}

func (s *authService) CompleteOAuth(ctx context.Context, input CompleteOAuthInput) (LoginResult, error) {
	provider, ok := s.oauth.Get(input.Provider)
	if !ok {
		return LoginResult{}, ErrUnsupportedOAuthProvider
	}

	if input.Code == "" {
		return LoginResult{}, ErrOAuthCodeRequired
	}
	if input.State == "" || input.BrowserState == "" || input.State != input.BrowserState {
		return LoginResult{}, ErrOAuthStateInvalid
	}

	attempt, err := s.getOAuthAttempt(ctx, input.State)
	if err != nil {
		return LoginResult{}, err
	}
	defer func() {
		_ = s.deleteOAuthAttempt(ctx, input.State)
	}()

	if attempt.Provider != provider.Name() {
		return LoginResult{}, ErrOAuthProviderMismatch
	}
	redirectURI := s.oauthCallbackURL(input.Provider)
	tokens, err := provider.ExchangeCode(ctx, OAuthExchangeCodeInput{
		RedirectURI:  redirectURI,
		Code:         input.Code,
		CodeVerifier: attempt.CodeVerifier,
	})
	if err != nil {
		return LoginResult{}, err
	}

	profile, err := provider.FetchUser(ctx, tokens, attempt.Nonce)
	if err != nil {
		return LoginResult{}, err
	}

	switch attempt.Mode {
	case "", OAuthModeLogin:
		return s.completeOAuthLogin(ctx, input, attempt, profile)
	case OAuthModeLink:
		return s.completeOAuthLink(ctx, attempt, profile)
	default:
		return LoginResult{}, ErrOAuthStateInvalid
	}
}

func (s *authService) completeOAuthLogin(
	ctx context.Context,
	input CompleteOAuthInput,
	attempt OAuthAttempt,
	profile OAuthProfile,
) (LoginResult, error) {
	user, err := s.repo.GetUserByExternalIdentity(ctx, profile.Provider, profile.ProviderUserID)
	if err == nil {
		result, innerErr := s.issueLoginResult(ctx, s.repo, user, input.UserAgent)
		if innerErr != nil {
			return LoginResult{}, innerErr
		}

		result.ReturnURL = attempt.ReturnURL
		return result, nil
	}
	if !errors.Is(err, authrepo.ErrNotFound) {
		return LoginResult{}, err
	}

	err = s.validateNewOAuthUser(ctx, profile)
	if err != nil {
		return LoginResult{}, err
	}

	var result LoginResult
	err = s.txManager.WithinTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	}, func(repositories *repository.Repositories) error {
		user, innerErr := repositories.Auth.CreateUser(ctx, authrepo.CreateUserInput{
			Username: profile.UsernameHint,
			Email:    profile.Email,
			Name:     profile.FirstName,
			Surname:  profile.LastName,
		})
		if innerErr != nil {
			return innerErr
		}

		_, innerErr = repositories.Auth.CreateExternalIdentity(ctx, authrepo.CreateExternalIdentityInput{
			UserID:           user.ID,
			Provider:         profile.Provider,
			ProviderSubject:  profile.ProviderUserID,
			ProviderUsername: stringPtrFromString(profile.UsernameHint),
		})
		if innerErr != nil {
			return innerErr
		}

		result, innerErr = s.issueLoginResult(ctx, repositories.Auth, user, input.UserAgent)
		return innerErr
	})
	if err != nil {
		var uniqueErr *authrepo.UniqueConstraintError
		if errors.As(err, &uniqueErr) {
			if containsString(uniqueErr.Fields, "email") {
				return LoginResult{}, ErrOAuthEmailAlreadyUsed
			}
			if containsString(uniqueErr.Fields, "username") {
				return LoginResult{}, ErrOAuthUsernameRequired
			}
		}

		return LoginResult{}, err
	}

	result.ReturnURL = attempt.ReturnURL
	return result, nil
}

func (s *authService) completeOAuthLink(ctx context.Context, attempt OAuthAttempt, profile OAuthProfile) (LoginResult, error) {
	if attempt.UserID == "" {
		return LoginResult{}, ErrOAuthStateInvalid
	}

	user, err := s.repo.GetUserByExternalIdentity(ctx, profile.Provider, profile.ProviderUserID)
	if err == nil {
		if user.ID == attempt.UserID {
			return LoginResult{ReturnURL: attempt.ReturnURL}, nil
		}

		return LoginResult{}, ErrOAuthIdentityAlreadyLinked
	}
	if !errors.Is(err, authrepo.ErrNotFound) {
		return LoginResult{}, err
	}

	_, err = s.repo.CreateExternalIdentity(ctx, authrepo.CreateExternalIdentityInput{
		UserID:           attempt.UserID,
		Provider:         profile.Provider,
		ProviderSubject:  profile.ProviderUserID,
		ProviderUsername: stringPtrFromString(profile.UsernameHint),
	})
	if err != nil {
		var uniqueErr *authrepo.UniqueConstraintError
		if errors.As(err, &uniqueErr) {
			if containsString(uniqueErr.Fields, "provider_subject") {
				return LoginResult{}, ErrOAuthIdentityAlreadyLinked
			}
			if containsString(uniqueErr.Fields, "provider") {
				return LoginResult{}, ErrOAuthProviderAlreadyLinked
			}
		}

		return LoginResult{}, err
	}

	return LoginResult{ReturnURL: attempt.ReturnURL}, nil
}

func (s *authService) startOAuthAttempt(
	ctx context.Context,
	providerName string,
	returnURL string,
	mode string,
	userID string,
) (StartOAuthResult, error) {
	provider, ok := s.oauth.Get(providerName)
	if !ok {
		return StartOAuthResult{}, ErrUnsupportedOAuthProvider
	}

	redirectURI := s.oauthCallbackURL(providerName)
	if redirectURI == "" {
		return StartOAuthResult{}, ErrOAuthRedirectURIRequired
	}

	state, err := randomURLSafe(32)
	if err != nil {
		return StartOAuthResult{}, err
	}
	codeVerifier, err := randomURLSafe(64)
	if err != nil {
		return StartOAuthResult{}, err
	}
	codeChallenge := pkceChallengeS256(codeVerifier)
	nonce, err := randomURLSafe(32)
	if err != nil {
		return StartOAuthResult{}, err
	}

	attempt := OAuthAttempt{
		Provider:     providerName,
		CodeVerifier: codeVerifier,
		Nonce:        nonce,
		ReturnURL:    s.safeOAuthReturnURL(returnURL),
		Mode:         mode,
		UserID:       userID,
	}

	redirectURL, err := provider.AuthCodeURL(OAuthAuthCodeURLInput{
		RedirectURI:   redirectURI,
		State:         state,
		CodeChallenge: codeChallenge,
		Nonce:         nonce,
		Scopes:        provider.Scopes(),
	})
	if err != nil {
		return StartOAuthResult{}, err
	}

	if err := s.saveOAuthAttempt(ctx, state, attempt); err != nil {
		return StartOAuthResult{}, err
	}

	return StartOAuthResult{
		RedirectURL: redirectURL,
		State:       state,
		MaxAge:      int(OAuthAttemptTTL / time.Second),
	}, nil
}

func (s *authService) saveOAuthAttempt(ctx context.Context, state string, attempt OAuthAttempt) error {
	data, err := json.Marshal(attempt)
	if err != nil {
		return err
	}

	return s.redisClient.Set(ctx, OAuthRedisKey+state, data, OAuthAttemptTTL).Err()
}

func (s *authService) getOAuthAttempt(ctx context.Context, state string) (OAuthAttempt, error) {
	data, err := s.redisClient.Get(ctx, OAuthRedisKey+state).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return OAuthAttempt{}, ErrOAuthStateInvalid
		}

		return OAuthAttempt{}, err
	}

	var attempt OAuthAttempt
	if err := json.Unmarshal(data, &attempt); err != nil {
		return OAuthAttempt{}, err
	}

	return attempt, nil
}

func (s *authService) deleteOAuthAttempt(ctx context.Context, state string) error {
	return s.redisClient.Del(ctx, OAuthRedisKey+state).Err()
}

func (s *authService) validateNewOAuthUser(ctx context.Context, profile OAuthProfile) error {
	if profile.UsernameHint == "" {
		return ErrOAuthUsernameRequired
	}

	conflicts, err := s.repo.FindUserConflicts(ctx, authrepo.FindUserConflictsInput{
		Username: profile.UsernameHint,
		Email:    profile.Email,
	})
	if err != nil {
		return err
	}

	if containsString(conflicts, "email") {
		return ErrOAuthEmailAlreadyUsed
	}
	if containsString(conflicts, "username") {
		return ErrOAuthUsernameRequired
	}

	return nil
}

func (s *authService) issueLoginResult(ctx context.Context, repo authRepository, user authrepo.User, userAgent string) (LoginResult, error) {
	tokenPair, err := s.tokens.IssueTokenPair(user.ID, userAgent)
	if err != nil {
		return LoginResult{}, err
	}

	refreshTokenHash := hashRefreshToken(tokenPair.RefreshToken)
	_, err = s.createRefreshSession(ctx, repo, authrepo.CreateRefreshSessionInput{
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		UserAgent: stringPtrFromString(userAgent),
		ExpiresAt: time.Now().Add(time.Duration(refreshTokenTTLSeconds) * time.Second),
	})
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		AccessToken:      tokenPair.AccessToken,
		RefreshToken:     tokenPair.RefreshToken,
		TokenType:        bearerTokenType,
		ExpiresIn:        accessTokenTTLSeconds,
		RefreshExpiresIn: refreshTokenTTLSeconds,
		User:             userFromRepository(user),
	}, nil
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}

	return false
}

func (s *authService) oauthCallbackURL(provider string) string {
	appPublicURL := strings.TrimRight(s.appPublicURL, "/")
	if appPublicURL == "" {
		return ""
	}

	return appPublicURL + "/api/v1/auth/" + provider + "/callback"
}

func (s *authService) safeOAuthReturnURL(rawReturnURL string) string {
	rawReturnURL = strings.TrimSpace(rawReturnURL)
	if rawReturnURL == "" {
		return "/"
	}

	returnURL, err := url.Parse(rawReturnURL)
	if err != nil {
		return "/"
	}

	if !returnURL.IsAbs() {
		if strings.HasPrefix(returnURL.Path, "/") && !strings.HasPrefix(rawReturnURL, "//") {
			return returnURL.String()
		}

		return "/"
	}

	appPublicURL, err := url.Parse(strings.TrimRight(s.appPublicURL, "/"))
	if err != nil || appPublicURL.Scheme == "" || appPublicURL.Host == "" {
		return "/"
	}

	if strings.EqualFold(returnURL.Scheme, appPublicURL.Scheme) &&
		strings.EqualFold(returnURL.Host, appPublicURL.Host) {
		returnURL.Scheme = ""
		returnURL.Host = ""
		returnURL.User = nil
		return returnURL.String()
	}

	return "/"
}

func randomURLSafe(byteLength int) (string, error) {
	b := make([]byte, byteLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func pkceChallengeS256(codeVerifier string) string {
	hash := sha256.Sum256([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
