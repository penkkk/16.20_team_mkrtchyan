package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	authrepo "opd/internal/repository/auth"

	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenTTLSeconds      = 15 * 60
	refreshTokenTTLSeconds     = 30 * 24 * 60 * 60
	refreshTokenHashByteLength = 32
)

func (s *service) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
	credentials, err := s.repo.GetUserCredentialsByLogin(ctx, input.Login)
	if err != nil {
		if errors.Is(err, authrepo.ErrNotFound) {
			return LoginResult{}, ErrInvalidCredentials
		}

		return LoginResult{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(credentials.PasswordHash), []byte(input.Password)); err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	tokenPair, err := s.tokens.IssueTokenPair(credentials.User.ID)
	if err != nil {
		return LoginResult{}, err
	}

	refreshTokenHash := hashRefreshToken(tokenPair.RefreshToken)
	_, err = s.repo.CreateRefreshSession(ctx, authrepo.CreateRefreshSessionInput{
		UserID:    credentials.User.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: time.Now().Add(time.Duration(refreshTokenTTLSeconds) * time.Second),
	})
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		AccessToken:      tokenPair.AccessToken,
		RefreshToken:     tokenPair.RefreshToken,
		TokenType:        "Bearer",
		ExpiresIn:        accessTokenTTLSeconds,
		RefreshExpiresIn: refreshTokenTTLSeconds,
		User:             userFromRepository(credentials.User),
	}, nil
}

func (s *service) Register(ctx context.Context, input RegisterInput) error {
	return ErrNotImplemented
}

func userFromRepository(user authrepo.User) User {
	return User{
		ID:         user.ID,
		Username:   user.Username,
		Email:      user.Email,
		TgUsername: user.TgUsername,
		Name:       user.Name,
		Surname:    user.Surname,
	}
}

func hashRefreshToken(refreshToken string) string {
	hash := sha256.Sum256([]byte(refreshToken))
	return hex.EncodeToString(hash[:refreshTokenHashByteLength])
}
