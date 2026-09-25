package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"opd/internal/repository"
	authrepo "opd/internal/repository/auth"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenTTLSeconds      = 15 * 60
	refreshTokenTTLSeconds     = 30 * 24 * 60 * 60
	refreshTokenHashByteLength = 32
)

func (s *authService) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
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

func (s *authService) Register(ctx context.Context, input RegisterInput) (LoginResult, error) {
	conflicts, err := s.repo.FindUserConflicts(ctx, authrepo.FindUserConflictsInput{
		Username:   input.Username,
		Email:      input.Email,
		TgUsername: input.TgUsername,
	})
	if err != nil {
		return LoginResult{}, err
	}
	if len(conflicts) > 0 {
		return LoginResult{}, NewFieldConflictError(conflicts)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return LoginResult{}, err
	}

	var result LoginResult
	err = s.txManager.WithinTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	}, func(repositories *repository.Repositories) error {
		user, err := repositories.Auth.CreateUser(ctx, authrepo.CreateUserInput{
			Username:   input.Username,
			Email:      input.Email,
			TgUsername: input.TgUsername,
			Name:       input.Name,
			Surname:    input.Surname,
		})
		if err != nil {
			return err
		}

		_, err = repositories.Auth.CreateUserPassword(ctx, authrepo.CreateUserPasswordInput{
			UserID:       user.ID,
			PasswordHash: string(passwordHash),
		})
		if err != nil {
			return err
		}

		tokenPair, err := s.tokens.IssueTokenPair(user.ID)
		if err != nil {
			return err
		}

		refreshTokenHash := hashRefreshToken(tokenPair.RefreshToken)
		_, err = repositories.Auth.CreateRefreshSession(ctx, authrepo.CreateRefreshSessionInput{
			UserID:    user.ID,
			TokenHash: refreshTokenHash,
			ExpiresAt: time.Now().Add(time.Duration(refreshTokenTTLSeconds) * time.Second),
		})
		if err != nil {
			return err
		}

		result = LoginResult{
			AccessToken:      tokenPair.AccessToken,
			RefreshToken:     tokenPair.RefreshToken,
			TokenType:        "Bearer",
			ExpiresIn:        accessTokenTTLSeconds,
			RefreshExpiresIn: refreshTokenTTLSeconds,
			User:             userFromRepository(user),
		}

		return nil
	})
	if err != nil {
		var uniqueErr *authrepo.UniqueConstraintError
		if errors.As(err, &uniqueErr) {
			return LoginResult{}, NewFieldConflictError(uniqueErr.Fields)
		}

		return LoginResult{}, err
	}

	return result, nil
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
