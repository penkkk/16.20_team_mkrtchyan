package tests

import (
	"strings"
	"testing"

	"opd/internal/config"
)

func TestJWTConfigurationRequiresAllValues(t *testing.T) {
	for _, tc := range []struct{ secret, issuer, audience, missing string }{
		{"", "issuer", "audience", "JWT_SECRET"},
		{"secret", "", "audience", "JWT_ISSUER"},
		{"secret", "issuer", "", "JWT_AUDIENCE"},
	} {
		if err := config.JWTSecretsCheck(tc.secret, tc.issuer, tc.audience); err == nil || !strings.Contains(err.Error(), tc.missing) {
			t.Fatalf("expected %s validation error, got %v", tc.missing, err)
		}
	}
	if err := config.JWTSecretsCheck("secret", "issuer", "audience"); err != nil {
		t.Fatal(err)
	}
}

func TestLoadConfiguration(t *testing.T) {
	t.Setenv("APP_PORT", "")
	t.Setenv("JWT_SECRET", " secret ")
	t.Setenv("JWT_ISSUER", " issuer ")
	t.Setenv("JWT_AUDIENCE", " audience ")
	t.Setenv("APP_PUBLIC_URL", " https://app.example.com/ ")
	loaded, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.HTTPPort != ":8080" || loaded.JWTSecret != "secret" || loaded.JWTIssuer != "issuer" || loaded.JWTAudience != "audience" || loaded.AppPublicURL != "https://app.example.com" {
		t.Fatalf("unexpected defaults or normalization: %+v", loaded)
	}
	t.Setenv("APP_PORT", ":9090")
	loaded, err = config.Load()
	if err != nil || loaded.HTTPPort != ":9090" {
		t.Fatalf("explicit port was not preserved: %+v, err=%v", loaded, err)
	}
	t.Setenv("JWT_SECRET", "")
	if _, err := config.Load(); err == nil {
		t.Fatal("missing JWT secret was accepted")
	}
}
