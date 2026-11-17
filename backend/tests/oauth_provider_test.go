package tests

import (
	"context"
	"errors"
	"net/url"
	"testing"

	authservice "opd/internal/service/auth"
)

func TestOAuthProviderAuthorizationURLs(t *testing.T) {
	for _, tc := range []struct {
		name, host string
		provider   authservice.OAuthProvider
	}{
		{"google", "accounts.google.com", authservice.NewGoogleOAuthProvider("client-id", "secret")},
		{"yandex", "oauth.yandex.ru", authservice.NewYandexOAuthProvider("client-id", "secret")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := authservice.OAuthAuthCodeURLInput{
				RedirectURI: "https://app.example.com/callback", State: "state", CodeChallenge: "challenge", Nonce: "nonce", Scopes: tc.provider.Scopes(),
			}
			rawURL, err := tc.provider.AuthCodeURL(input)
			if err != nil {
				t.Fatal(err)
			}
			u, err := url.Parse(rawURL)
			if err != nil {
				t.Fatal(err)
			}
			q := u.Query()
			if u.Scheme != "https" || u.Host != tc.host || q.Get("client_id") != "client-id" || q.Get("redirect_uri") != input.RedirectURI ||
				q.Get("response_type") != "code" || q.Get("state") != "state" || q.Get("code_challenge") != "challenge" || q.Get("code_challenge_method") != "S256" || q.Get("scope") == "" {
				t.Fatalf("invalid authorization URL: %s", rawURL)
			}
			if tc.name == "google" && q.Get("nonce") != "nonce" {
				t.Fatalf("google nonce missing: %s", rawURL)
			}
		})
	}
}

func TestOAuthProviderRejectsMissingConfiguration(t *testing.T) {
	for _, provider := range []authservice.OAuthProvider{
		authservice.NewGoogleOAuthProvider("", ""),
		authservice.NewYandexOAuthProvider("", ""),
	} {
		t.Run(provider.Name(), func(t *testing.T) {
			if _, err := provider.AuthCodeURL(authservice.OAuthAuthCodeURLInput{RedirectURI: "https://app.example.com/callback"}); !errors.Is(err, authservice.ErrOAuthClientIDRequired) {
				t.Fatalf("missing client ID: %v", err)
			}
			if _, err := provider.FetchUser(context.Background(), authservice.OAuthTokens{}, ""); !errors.Is(err, authservice.ErrInvalidAccessToken) {
				t.Fatalf("missing OAuth token: %v", err)
			}
		})
	}
}
