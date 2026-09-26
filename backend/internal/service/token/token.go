package token

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Manager interface {
	IssueToken(userID string, userAgent string) (string, error)
	IssueTokenPair(userID string, userAgent string) (TokenPair, error)
	VerifyAccessToken(tokenString string) (Claims, error)
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type Claims struct {
	UserID    string
	UserAgent string
}

type accessTokenClaims struct {
	UserAgent string `json:"user_agent,omitempty"`

	jwt.RegisteredClaims
}

type JWTManager struct {
	secret    string
	issuer    string
	audience  string
	accessTTL time.Duration
}

func NewJWTManager(secret, issuer, audience string, accessTTL time.Duration) *JWTManager {
	return &JWTManager{
		secret:    secret,
		issuer:    issuer,
		audience:  audience,
		accessTTL: accessTTL,
	}
}

func (m *JWTManager) IssueToken(userID string, userAgent string) (string, error) {
	now := time.Now()
	claims := accessTokenClaims{
		UserAgent: userAgent,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Audience:  jwt.ClaimStrings{m.audience},
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}

func (m *JWTManager) IssueTokenPair(userID string, userAgent string) (TokenPair, error) {
	accessToken, err := m.IssueToken(userID, userAgent)
	if err != nil {
		return TokenPair{}, err
	}

	refreshToken, err := randomToken(32)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (m *JWTManager) VerifyAccessToken(tokenString string) (Claims, error) {
	claims := accessTokenClaims{}

	parsedToken, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}

			return []byte(m.secret), nil
		},
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(m.issuer),
		jwt.WithAudience(m.audience),
	)
	if err != nil {
		return Claims{}, err
	}
	if !parsedToken.Valid {
		return Claims{}, errors.New("invalid token")
	}

	return Claims{
		UserID:    claims.Subject,
		UserAgent: claims.UserAgent,
	}, nil
}

func randomToken(size int) (string, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
