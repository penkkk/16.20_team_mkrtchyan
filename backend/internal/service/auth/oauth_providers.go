package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"google.golang.org/api/idtoken"
)

const (
	googleOAuthAuthorizeURL = "https://accounts.google.com/o/oauth2/v2/auth"
	yandexOAuthAuthorizeURL = "https://oauth.yandex.ru/authorize"
	googleTokenURL          = "https://oauth2.googleapis.com/token"
	yandexTokenURL          = "https://oauth.yandex.ru/token"
	yandexUserInfoURL       = "https://login.yandex.ru/info?format=json"
)

type GoogleOAuthProvider struct {
	clientID     string
	clientSecret string
}

func NewGoogleOAuthProvider(clientID string, clientSecret string) *GoogleOAuthProvider {
	return &GoogleOAuthProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func (p *GoogleOAuthProvider) Name() string {
	return OAuthProviderGoogle
}

func (p *GoogleOAuthProvider) Scopes() []string {
	return []string{"openid", "email", "profile"}
}

func (p *GoogleOAuthProvider) AuthCodeURL(input OAuthAuthCodeURLInput) (string, error) {
	if p.clientID == "" {
		return "", ErrOAuthClientIDRequired
	}
	if input.RedirectURI == "" {
		return "", ErrOAuthRedirectURIRequired
	}

	u, err := url.Parse(googleOAuthAuthorizeURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("client_id", p.clientID)
	q.Set("redirect_uri", input.RedirectURI)
	q.Set("response_type", "code")
	q.Set("scope", strings.Join(input.Scopes, " "))
	q.Set("state", input.State)
	q.Set("code_challenge", input.CodeChallenge)
	q.Set("code_challenge_method", "S256")
	if input.Nonce != "" {
		q.Set("nonce", input.Nonce)
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (p *GoogleOAuthProvider) ExchangeCode(ctx context.Context, input OAuthExchangeCodeInput) (OAuthTokens, error) {
	return exchangeOAuthCode(ctx, googleTokenURL, OAuthProviderGoogle, p.clientID, p.clientSecret, input)
}

func (p *GoogleOAuthProvider) FetchUser(ctx context.Context, tokens OAuthTokens, expectedNonce string) (OAuthProfile, error) {
	if tokens.IDToken == "" {
		return OAuthProfile{}, ErrInvalidAccessToken
	}

	payload, err := idtoken.Validate(ctx, tokens.IDToken, p.clientID)
	if err != nil {
		return OAuthProfile{}, err
	}

	email, _ := payload.Claims["email"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)
	firstName, _ := payload.Claims["given_name"].(string)
	lastName, _ := payload.Claims["family_name"].(string)
	nonce, _ := payload.Claims["nonce"].(string)
	if expectedNonce != "" && nonce != expectedNonce {
		return OAuthProfile{}, ErrOAuthNonceMismatch
	}

	username, _, ok := strings.Cut(email, "@")
	if !ok {
		return OAuthProfile{}, ErrInvalidCredentials
	}

	return OAuthProfile{
		Provider:       OAuthProviderGoogle,
		ProviderUserID: payload.Subject,
		Email:          email,
		EmailVerified:  emailVerified,
		FirstName:      firstName,
		LastName:       lastName,
		UsernameHint:   username,
	}, nil
}

type YandexOAuthProvider struct {
	clientID     string
	clientSecret string
}

func NewYandexOAuthProvider(clientID string, clientSecret string) *YandexOAuthProvider {
	return &YandexOAuthProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func (p *YandexOAuthProvider) Name() string {
	return OAuthProviderYandex
}

func (p *YandexOAuthProvider) Scopes() []string {
	return []string{"login:info", "login:email"}
}

func (p *YandexOAuthProvider) AuthCodeURL(input OAuthAuthCodeURLInput) (string, error) {
	if p.clientID == "" {
		return "", ErrOAuthClientIDRequired
	}
	if input.RedirectURI == "" {
		return "", ErrOAuthRedirectURIRequired
	}

	u, err := url.Parse(yandexOAuthAuthorizeURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("client_id", p.clientID)
	q.Set("redirect_uri", input.RedirectURI)
	q.Set("response_type", "code")
	q.Set("scope", strings.Join(input.Scopes, " "))
	q.Set("state", input.State)
	q.Set("code_challenge", input.CodeChallenge)
	q.Set("code_challenge_method", "S256")

	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (p *YandexOAuthProvider) ExchangeCode(ctx context.Context, input OAuthExchangeCodeInput) (OAuthTokens, error) {
	return exchangeOAuthCode(ctx, yandexTokenURL, OAuthProviderYandex, p.clientID, p.clientSecret, input)
}

func (p *YandexOAuthProvider) FetchUser(ctx context.Context, tokens OAuthTokens, _ string) (OAuthProfile, error) {
	if tokens.AccessToken == "" {
		return OAuthProfile{}, ErrInvalidAccessToken
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		yandexUserInfoURL,
		nil,
	)
	if err != nil {
		return OAuthProfile{}, err
	}

	req.Header.Set("Authorization", "OAuth "+tokens.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return OAuthProfile{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return OAuthProfile{}, fmt.Errorf("yandex user info failed: %s", resp.Status)
	}

	var userInfo struct {
		ID           string `json:"id"`
		Login        string `json:"login"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		DefaultEmail string `json:"default_email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return OAuthProfile{}, err
	}

	return OAuthProfile{
		Provider:       OAuthProviderYandex,
		ProviderUserID: userInfo.ID,
		Email:          userInfo.DefaultEmail,
		EmailVerified:  userInfo.DefaultEmail != "",
		FirstName:      userInfo.FirstName,
		LastName:       userInfo.LastName,
		UsernameHint:   userInfo.Login,
	}, nil
}

func exchangeOAuthCode(
	ctx context.Context,
	tokenURL string,
	providerName string,
	clientID string,
	clientSecret string,
	input OAuthExchangeCodeInput,
) (OAuthTokens, error) {
	if clientID == "" {
		return OAuthTokens{}, ErrOAuthClientIDRequired
	}
	if clientSecret == "" {
		return OAuthTokens{}, ErrOAuthClientSecretRequired
	}
	if input.RedirectURI == "" {
		return OAuthTokens{}, ErrOAuthRedirectURIRequired
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("code", input.Code)
	form.Set("redirect_uri", input.RedirectURI)
	form.Set("code_verifier", input.CodeVerifier)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		tokenURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return OAuthTokens{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return OAuthTokens{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return OAuthTokens{}, fmt.Errorf("%s oauth token exchange failed: %s", providerName, resp.Status)
	}

	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return OAuthTokens{}, err
	}

	return OAuthTokens{
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: tokenResponse.RefreshToken,
		IDToken:      tokenResponse.IDToken,
		TokenType:    tokenResponse.TokenType,
		ExpiresIn:    tokenResponse.ExpiresIn,
	}, nil
}

func verifyGoogleIDToken(ctx context.Context, rawIDToken string, clientID string) (*idtoken.Payload, error) {
	payload, err := idtoken.Validate(ctx, rawIDToken, clientID)
	if err != nil {
		return nil, err
	}

	return payload, nil
}
