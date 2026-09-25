package auth

import (
	"context"
	"time"

	"opd/internal/db"

	"github.com/jackc/pgx/v5/pgtype"
)

func (r *PostgresRepository) CreateRefreshSession(ctx context.Context, input CreateRefreshSessionInput) (RefreshSession, error) {
	userID, err := uuidFromString(input.UserID)
	if err != nil {
		return RefreshSession{}, err
	}

	session, err := r.queries.CreateRefreshSession(ctx, db.CreateRefreshSessionParams{
		UserID:    userID,
		TokenHash: input.TokenHash,
		UserAgent: textFromStringPtr(input.UserAgent),
		Ip:        input.IP,
		ExpiresAt: timestamptzFromTime(input.ExpiresAt),
	})
	if err != nil {
		return RefreshSession{}, err
	}

	return refreshSessionFromDB(session), nil
}

func (r *PostgresRepository) RevokeRefreshSession(ctx context.Context, tokenHash string) (RefreshSession, error) {
	session, err := r.queries.RevokeRefreshSession(ctx, tokenHash)
	if err != nil {
		return RefreshSession{}, err
	}

	return refreshSessionFromDB(session), nil
}

func refreshSessionFromDB(session db.AuthRefreshSession) RefreshSession {
	return RefreshSession{
		ID:        session.ID.String(),
		UserID:    session.UserID.String(),
		TokenHash: session.TokenHash,
		UserAgent: stringFromText(session.UserAgent),
		IP:        session.Ip,
		ExpiresAt: session.ExpiresAt.Time,
		RevokedAt: timePtrFromTimestamptz(session.RevokedAt),
		CreatedAt: session.CreatedAt.Time,
	}
}

func timestamptzFromTime(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  value,
		Valid: true,
	}
}

func timePtrFromTimestamptz(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	return &value.Time
}
