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
	bearerTokenType            = "Bearer"
	redisBlackListPrefix       = "auth:blacklist:"
)

func (s *authService) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
	credentials, err := s.repo.GetUserCredentialsByLogin(ctx, input.Login)
	if err != nil {
		if errors.Is(err, authrepo.ErrNotFound) {
			return LoginResult{}, ErrInvalidCredentials
		}

		return LoginResult{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(credentials.PasswordHash), []byte(input.Password))
	if err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	tokenPair, err := s.tokens.IssueTokenPair(credentials.User.ID, input.UserAgent)
	if err != nil {
		return LoginResult{}, err
	}

	refreshTokenHash := hashRefreshToken(tokenPair.RefreshToken)
	_, err = s.createRefreshSession(ctx, s.repo, authrepo.CreateRefreshSessionInput{
		UserID:    credentials.User.ID,
		TokenHash: refreshTokenHash,
		UserAgent: stringPtrFromString(input.UserAgent),
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
		user, innerErr := repositories.Auth.CreateUser(ctx, authrepo.CreateUserInput{
			Username:   input.Username,
			Email:      input.Email,
			TgUsername: input.TgUsername,
			Name:       input.Name,
			Surname:    input.Surname,
		})
		if innerErr != nil {
			return innerErr
		}

		_, innerErr = repositories.Auth.CreateUserPassword(ctx, authrepo.CreateUserPasswordInput{
			UserID:       user.ID,
			PasswordHash: string(passwordHash),
		})
		if innerErr != nil {
			return innerErr
		}

		tokenPair, innerErr := s.tokens.IssueTokenPair(user.ID, input.UserAgent)
		if innerErr != nil {
			return innerErr
		}

		refreshTokenHash := hashRefreshToken(tokenPair.RefreshToken)
		_, innerErr = s.createRefreshSession(ctx, repositories.Auth, authrepo.CreateRefreshSessionInput{
			UserID:    user.ID,
			TokenHash: refreshTokenHash,
			UserAgent: stringPtrFromString(input.UserAgent),
			ExpiresAt: time.Now().Add(time.Duration(refreshTokenTTLSeconds) * time.Second),
		})
		if innerErr != nil {
			return innerErr
		}

		result = LoginResult{
			AccessToken:      tokenPair.AccessToken,
			RefreshToken:     tokenPair.RefreshToken,
			TokenType:        bearerTokenType,
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

func (s *authService) Logout(ctx context.Context, input LogoutInput) error {
	refreshTokenHash := hashRefreshToken(input.RefreshSession)
	_, err := s.repo.RevokeRefreshSession(ctx, refreshTokenHash)
	if err != nil {
		if errors.Is(err, authrepo.ErrNotFound) {
			return ErrInvalidRefreshSession
		}

		return err
	}

	return s.addAccessTokenToBlacklist(ctx, input.AccessToken)
}

func (s *authService) Refresh(ctx context.Context, input LogoutInput) (RefreshResult, error) {
	oldRefreshTokenHash := hashRefreshToken(input.RefreshSession)
	var result RefreshResult
	err := s.txManager.WithinTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	}, func(repositories *repository.Repositories) error {
		oldSession, err := repositories.Auth.RevokeRefreshSession(ctx, oldRefreshTokenHash)
		if err != nil {
			if errors.Is(err, authrepo.ErrNotFound) {
				return ErrInvalidRefreshSession
			}

			return err
		}

		tokenPair, err := s.tokens.IssueTokenPair(oldSession.UserID, oldSession.UserAgent)
		if err != nil {
			return err
		}

		newRefreshTokenHash := hashRefreshToken(tokenPair.RefreshToken)
		refreshSession := authrepo.CreateRefreshSessionInput{
			UserID:    oldSession.UserID,
			UserAgent: stringPtrFromString(oldSession.UserAgent),
			TokenHash: newRefreshTokenHash,
			IP:        nil,
			ExpiresAt: time.Now().Add(time.Duration(refreshTokenTTLSeconds) * time.Second),
		}

		_, err = s.createRefreshSession(ctx, repositories.Auth, refreshSession)
		if err != nil {
			return err
		}
		result = RefreshResult{
			AccessToken:      tokenPair.AccessToken,
			RefreshToken:     tokenPair.RefreshToken,
			TokenType:        bearerTokenType,
			ExpiresIn:        accessTokenTTLSeconds,
			RefreshExpiresIn: refreshTokenTTLSeconds,
		}
		return nil
	})
	if err != nil {
		return RefreshResult{}, err
	}

	if err := s.addAccessTokenToBlacklist(ctx, input.AccessToken); err != nil {
		return RefreshResult{}, err
	}

	return result, nil
}

func (s *authService) AddPassword(ctx context.Context, input AddPasswordInput) error {
	err := s.repo.GetUserPasswordByID(ctx, authrepo.GetPasswordByUserIDInput{
		UserID: input.UserID,
	})
	if err == nil {
		return ErrPasswordAlreadySet
	}
	if !errors.Is(err, authrepo.ErrNotFound) {
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.repo.CreateUserPassword(ctx, authrepo.CreateUserPasswordInput{
		UserID:       input.UserID,
		PasswordHash: string(passwordHash),
	})
	if err != nil {
		var uniqueErr *authrepo.UniqueConstraintError
		if errors.As(err, &uniqueErr) {
			return ErrPasswordAlreadySet
		}

		return err
	}

	return nil
}

func (s *authService) createRefreshSession(
	ctx context.Context,
	repo authRepository,
	input authrepo.CreateRefreshSessionInput,
) (authrepo.RefreshSession, error) {
	if err := repo.RevokeActiveRefreshSessionsByUserAgent(ctx, input.UserID, input.UserAgent); err != nil {
		return authrepo.RefreshSession{}, err
	}

	return repo.CreateRefreshSession(ctx, input)
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

func (s *authService) addAccessTokenToBlacklist(ctx context.Context, accessToken string) error {
	if accessToken == "" {
		return nil
	}

	claims, err := s.tokens.VerifyAccessToken(accessToken)
	if err != nil {
		return nil
	}

	ttl := time.Until(claims.ExpiresAt)
	if ttl <= 0 {
		return nil
	}

	key := redisBlackListPrefix + hashToken(accessToken)

	return s.redisClient.Set(ctx, key, "1", ttl).Err()
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func stringPtrFromString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}
