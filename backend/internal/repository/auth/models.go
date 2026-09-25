package auth

import (
	"net/netip"
	"time"
)

type User struct {
	ID         string
	Username   string
	Email      string
	TgUsername string
	Name       string
	Surname    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CreateUserInput struct {
	Username   string
	Email      string
	TgUsername *string
	Name       string
	Surname    string
}

type UserPassword struct {
	ID           string
	UserID       string
	PasswordHash string
}

type UserCredentials struct {
	User         User
	PasswordHash string
}

type CreateUserPasswordInput struct {
	UserID       string
	PasswordHash string
}

type RefreshSession struct {
	ID        string
	UserID    string
	TokenHash string
	UserAgent string
	IP        *netip.Addr
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type CreateRefreshSessionInput struct {
	UserID    string
	TokenHash string
	UserAgent *string
	IP        *netip.Addr
	ExpiresAt time.Time
}
