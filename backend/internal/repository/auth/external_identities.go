package auth

import (
	"context"
	"errors"

	"opd/internal/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *AuthPostgresRepository) GetUserByExternalIdentity(ctx context.Context, provider string, providerSubject string) (User, error) {
	user, err := r.queries.GetUserByExternalIdentity(ctx, db.GetUserByExternalIdentityParams{
		Provider:        provider,
		ProviderSubject: providerSubject,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}

		return User{}, err
	}

	return userFromDB(user), nil
}

func (r *AuthPostgresRepository) CreateExternalIdentity(ctx context.Context, input CreateExternalIdentityInput) (ExternalIdentity, error) {
	userID, err := uuidFromString(input.UserID)
	if err != nil {
		return ExternalIdentity{}, err
	}

	identity, err := r.queries.CreateExternalIdentity(ctx, db.CreateExternalIdentityParams{
		UserID:           userID,
		Provider:         input.Provider,
		ProviderSubject:  input.ProviderSubject,
		ProviderUsername: textFromStringPtr(input.ProviderUsername),
	})
	if err != nil {
		if fields := externalIdentityUniqueConstraintFields(err); len(fields) > 0 {
			return ExternalIdentity{}, NewUniqueConstraintError(fields...)
		}

		return ExternalIdentity{}, err
	}

	return ExternalIdentity{
		ID:               identity.ID.String(),
		UserID:           identity.UserID.String(),
		Provider:         identity.Provider,
		ProviderSubject:  identity.ProviderSubject,
		ProviderUsername: stringFromText(identity.ProviderUsername),
		CreatedAt:        identity.CreatedAt.Time,
		UpdatedAt:        identity.UpdatedAt.Time,
	}, nil
}

func externalIdentityUniqueConstraintFields(err error) []string {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != uniqueViolationCode {
		return nil
	}

	switch pgErr.ConstraintName {
	case "external_identities_provider_provider_subject_key":
		return []string{"provider_subject"}
	case "external_identities_user_id_provider_key":
		return []string{"provider"}
	default:
		return nil
	}
}
