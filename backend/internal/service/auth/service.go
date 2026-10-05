package auth

import (
	"context"

	"opd/internal/repository"
	authrepo "opd/internal/repository/auth"
	token "opd/internal/service/token"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type AuthService interface {
	Login(ctx context.Context, input LoginInput) (LoginResult, error)
	Register(ctx context.Context, input RegisterInput) (LoginResult, error)
	Logout(ctx context.Context, input LogoutInput) error
	Refresh(ctx context.Context, input LogoutInput) (RefreshResult, error)
	StartOAuth(ctx context.Context, input StartOAuthInput) (StartOAuthResult, error)
	CompleteOAuth(ctx context.Context, input CompleteOAuthInput) (CompleteOAuthResult, error)
	CompleteOAuthRegistration(ctx context.Context, input CompleteOAuthRegistrationInput) (LoginResult, error)
	StartOAuthLink(ctx context.Context, input StartOAuthLinkInput) (StartOAuthResult, error)
	AddPassword(ctx context.Context, input AddPasswordInput) error
	Me(ctx context.Context, input MeInput) (User, error)
}

type authRepository interface {
	CreateUser(ctx context.Context, input authrepo.CreateUserInput) (authrepo.User, error)
	CreateUserPassword(ctx context.Context, input authrepo.CreateUserPasswordInput) (authrepo.UserPassword, error)
	GetUserByID(ctx context.Context, input authrepo.GetUserByIDInput) (authrepo.User, error)
	GetUserByExternalIdentity(ctx context.Context, provider string, providerSubject string) (authrepo.User, error)
	CreateExternalIdentity(ctx context.Context, input authrepo.CreateExternalIdentityInput) (authrepo.ExternalIdentity, error)
	GetUserCredentialsByLogin(ctx context.Context, login string) (authrepo.UserCredentials, error)
	FindUserConflicts(ctx context.Context, input authrepo.FindUserConflictsInput) ([]string, error)
	CreateRefreshSession(ctx context.Context, input authrepo.CreateRefreshSessionInput) (authrepo.RefreshSession, error)
	RevokeRefreshSession(ctx context.Context, tokenHash string) (authrepo.RevokedRefreshSession, error)
	RevokeActiveRefreshSessionsByUserAgent(ctx context.Context, userID string, userAgent *string) error
	GetUserPasswordByID(ctx context.Context, input authrepo.GetPasswordByUserIDInput) error
}

type authTxManager interface {
	WithinTx(ctx context.Context, opts pgx.TxOptions, fn func(*repository.Repositories) error) error
}

type authService struct {
	repo         authRepository
	txManager    authTxManager
	tokens       token.Manager
	redisClient  *redis.Client
	oauth        *OAuthProviderRegistry
	appPublicURL string
}

type OAuthConfig struct {
	GoogleClientID     string
	YandexClientID     string
	GoogleClientSecret string
	YandexClientSecret string
	AppPublicURL       string
}

func NewAuthService(
	repo authRepository,
	txManager authTxManager,
	tokens token.Manager,
	redisClient *redis.Client,
	oauthConfig OAuthConfig,
) AuthService {
	return &authService{
		repo:         repo,
		txManager:    txManager,
		tokens:       tokens,
		redisClient:  redisClient,
		appPublicURL: oauthConfig.AppPublicURL,
		oauth: NewOAuthProviderRegistry(
			NewGoogleOAuthProvider(oauthConfig.GoogleClientID, oauthConfig.GoogleClientSecret),
			NewYandexOAuthProvider(oauthConfig.YandexClientID, oauthConfig.YandexClientSecret),
		),
	}
}
