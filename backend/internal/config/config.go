package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	HTTPPort    string
	DataBaseURL string
	JWTSecret   string
	JWTIssuer   string
	JWTAudience string
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
		HTTPPort:    httpPort,
		DataBaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   jwtSecret,
		JWTIssuer:   jwtIssuer,
		JWTAudience: jwtAudience,
	}, nil
}


func JWTSecretsCheck(jwtSecret string, jwtIssuer string, jwtAudience string) (error){
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