package tests

import (
	"context"
	"errors"
	"testing"

	authservice "opd/internal/service/auth"
)

func TestOAuthRejectsInvalidRequests(t *testing.T) {
	service := authservice.NewAuthService(nil, nil, nil, nil, authservice.OAuthConfig{AppPublicURL: "https://app.example.com"})
	ctx := context.Background()
	if _, err := service.StartOAuth(ctx, authservice.StartOAuthInput{Provider: "unknown"}); !errors.Is(err, authservice.ErrUnsupportedOAuthProvider) {
		t.Fatalf("unsupported provider: %v", err)
	}
	if _, err := service.StartOAuthLink(ctx, authservice.StartOAuthLinkInput{Provider: "google"}); !errors.Is(err, authservice.ErrInvalidCredentials) {
		t.Fatalf("link without user: %v", err)
	}
	if _, err := service.CompleteOAuth(ctx, authservice.CompleteOAuthInput{Provider: "google"}); !errors.Is(err, authservice.ErrOAuthCodeRequired) {
		t.Fatalf("callback without code: %v", err)
	}
	if _, err := service.CompleteOAuth(ctx, authservice.CompleteOAuthInput{Provider: "google", Code: "code", State: "state", BrowserState: "other"}); !errors.Is(err, authservice.ErrOAuthStateInvalid) {
		t.Fatalf("callback with mismatched state: %v", err)
	}
}
