package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	HTTPPort                string
	AppPublicURL            string
	DataBaseURL             string
	JWTSecret               string
	JWTIssuer               string
	JWTAudience             string
	RedisURL                string
	GoogleOAuthClientID     string
	YandexOAuthClientID     string
	YandexOAuthClientSecret string
	GoogleOAuthClientSecret string
}

func Load() (Config, error) {
	httpPort := os.Getenv("APP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	if !strings.HasPrefix(httpPort, ":") {
		httpPort = ":" + httpPort
	}

	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	jwtIssuer := strings.TrimSpace(os.Getenv("JWT_ISSUER"))
	jwtAudience := strings.TrimSpace(os.Getenv("JWT_AUDIENCE"))

	err := JWTSecretsCheck(jwtSecret, jwtIssuer, jwtAudience)
	if err != nil {
		return Config{}, err
	}

	return Config{
		HTTPPort:                httpPort,
		AppPublicURL:            strings.TrimRight(strings.TrimSpace(os.Getenv("APP_PUBLIC_URL")), "/"),
		DataBaseURL:             os.Getenv("DATABASE_URL"),
		RedisURL:                os.Getenv("REDIS_URL"),
		JWTSecret:               jwtSecret,
		JWTIssuer:               jwtIssuer,
		JWTAudience:             jwtAudience,
		GoogleOAuthClientID:     strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID")),
		YandexOAuthClientID:     strings.TrimSpace(os.Getenv("YANDEX_CLIENT_ID")),
		GoogleOAuthClientSecret: strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET")),
		YandexOAuthClientSecret: strings.TrimSpace(os.Getenv("YANDEX_CLIENT_SECRET")),
	}, nil
}

func JWTSecretsCheck(jwtSecret string, jwtIssuer string, jwtAudience string) error {
	if jwtSecret == "" {
		return errors.New("JWT_SECRET is required")
	}
	if jwtIssuer == "" {
		return errors.New("JWT_ISSUER is required")
	}
	if jwtAudience == "" {
		return errors.New("JWT_AUDIENCE is required")
	}

	return nil
}
