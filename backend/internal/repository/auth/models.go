package auth

import (
	"net/netip"
	"time"
)

type User struct {
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ID         string
	Username   string
	Email      string
	TgUsername string
	Name       string
	Surname    string
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
	ExpiresAt time.Time
	CreatedAt time.Time
	IP        *netip.Addr
	RevokedAt *time.Time
	ID        string
	UserID    string
	TokenHash string
	UserAgent string
}

type CreateRefreshSessionInput struct {
	ExpiresAt time.Time
	UserAgent *string
	IP        *netip.Addr
	UserID    string
	TokenHash string
}
