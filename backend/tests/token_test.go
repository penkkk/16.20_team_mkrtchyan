package tests

import (
	"encoding/base64"
	"testing"
	"time"

	"opd/internal/service/token"
)

func TestIssueTokenPair(t *testing.T) {
	manager := token.NewJWTManager("secret", "issuer", "audience", time.Minute)
	first, err := manager.IssueTokenPair("user-1", "browser")
	if err != nil {
		t.Fatal(err)
	}
	second, err := manager.IssueTokenPair("user-1", "browser")
	if err != nil {
		t.Fatal(err)
	}
	if first.RefreshToken == second.RefreshToken || first.AccessToken == "" {
		t.Fatal("expected a signed access token and unique refresh tokens")
	}
	refreshBytes, err := base64.RawURLEncoding.DecodeString(first.RefreshToken)
	if err != nil || len(refreshBytes) != 32 {
		t.Fatalf("refresh token must contain 32 random bytes: length=%d, err=%v", len(refreshBytes), err)
	}
	claims, err := manager.VerifyAccessToken(first.AccessToken)
	if err != nil || claims.UserID != "user-1" || claims.UserAgent != "browser" {
		t.Fatalf("unexpected claims: %+v, err=%v", claims, err)
	}
	if remaining := time.Until(claims.ExpiresAt); remaining < 58*time.Second || remaining > time.Minute {
		t.Fatalf("unexpected expiry: %v", remaining)
	}
}

func TestVerifyAccessTokenRejectsInvalidTokens(t *testing.T) {
	manager := token.NewJWTManager("secret", "issuer", "audience", time.Minute)
	valid, err := manager.IssueToken("user-1", "browser")
	if err != nil {
		t.Fatal(err)
	}
	for name, verifier := range map[string]*token.JWTManager{
		"secret":   token.NewJWTManager("other", "issuer", "audience", time.Minute),
		"issuer":   token.NewJWTManager("secret", "other", "audience", time.Minute),
		"audience": token.NewJWTManager("secret", "issuer", "other", time.Minute),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := verifier.VerifyAccessToken(valid); err == nil {
				t.Fatal("expected verification error")
			}
		})
	}
	expired, err := token.NewJWTManager("secret", "issuer", "audience", -time.Minute).IssueToken("user-1", "browser")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.VerifyAccessToken(expired); err == nil {
		t.Fatal("expected expired token to be rejected")
	}
}
