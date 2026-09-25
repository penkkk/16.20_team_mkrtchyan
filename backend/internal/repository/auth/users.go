package auth

import (
	"context"
	"errors"

	"opd/internal/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *PostgresRepository) CreateUser(ctx context.Context, input CreateUserInput) (User, error) {
	user, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Username:   input.Username,
		Email:      input.Email,
		TgUsername: textFromStringPtr(input.TgUsername),
		Name:       textFromString(input.Name),
		Surname:    textFromString(input.Surname),
	})
	if err != nil {
		return User{}, err
	}

	return userFromDB(user), nil
}

func (r *PostgresRepository) CreateUserPassword(ctx context.Context, input CreateUserPasswordInput) (UserPassword, error) {
	userID, err := uuidFromString(input.UserID)
	if err != nil {
		return UserPassword{}, err
	}

	password, err := r.queries.CreateUserPassword(ctx, db.CreateUserPasswordParams{
		UserID:       userID,
		PasswordHash: input.PasswordHash,
	})
	if err != nil {
		return UserPassword{}, err
	}

	return userPasswordFromDB(password), nil
}

func (r *PostgresRepository) GetUserCredentialsByLogin(ctx context.Context, login string) (UserCredentials, error) {
	credentials, err := r.queries.GetUserCredentialsByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserCredentials{}, ErrNotFound
		}

		return UserCredentials{}, err
	}

	return userCredentialsFromDB(credentials), nil
}

func userFromDB(user db.AuthUser) User {
	return User{
		ID:         user.ID.String(),
		Username:   user.Username,
		Email:      user.Email,
		TgUsername: stringFromText(user.TgUsername),
		Name:       stringFromText(user.Name),
		Surname:    stringFromText(user.Surname),
		CreatedAt:  user.CreatedAt.Time,
		UpdatedAt:  user.UpdatedAt.Time,
	}
}

func userCredentialsFromDB(credentials db.GetUserCredentialsByLoginRow) UserCredentials {
	return UserCredentials{
		User: User{
			ID:         credentials.ID.String(),
			Username:   credentials.Username,
			Email:      credentials.Email,
			TgUsername: stringFromText(credentials.TgUsername),
			Name:       stringFromText(credentials.Name),
			Surname:    stringFromText(credentials.Surname),
			CreatedAt:  credentials.CreatedAt.Time,
			UpdatedAt:  credentials.UpdatedAt.Time,
		},
		PasswordHash: credentials.PasswordHash,
	}
}

func userPasswordFromDB(password db.AuthUserPassword) UserPassword {
	return UserPassword{
		ID:           password.ID.String(),
		UserID:       password.UserID.String(),
		PasswordHash: password.PasswordHash,
	}
}

func uuidFromString(value string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	if err := uuid.Scan(value); err != nil {
		return pgtype.UUID{}, err
	}

	return uuid, nil
}

func textFromStringPtr(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}

	return textFromString(*value)
}

func textFromString(value string) pgtype.Text {
	return pgtype.Text{
		String: value,
		Valid:  true,
	}
}

func stringFromText(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}

	return value.String
}
