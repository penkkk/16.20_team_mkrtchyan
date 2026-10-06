package auth

type LoginInput struct {
	Login     string
	Password  string
	UserAgent string
}

type LoginResult struct {
	User             User
	AccessToken      string
	RefreshToken     string
	ReturnURL        string
	TokenType        string
	ExpiresIn        int
	RefreshExpiresIn int
}

type CompleteOAuthResult struct {
	PendingToken string
	LoginResult
}

type RefreshResult struct {
	AccessToken      string
	RefreshToken     string
	TokenType        string
	ExpiresIn        int
	RefreshExpiresIn int
}

type RegisterInput struct {
	Username   string
	Email      string
	Password   string
	TgUsername *string
	Name       string
	Surname    string
	UserAgent  string
}

type LogoutInput struct {
	AccessToken    string
	RefreshSession string
}

type StartOAuthInput struct {
	Provider  string
	ReturnURL string
}

type StartOAuthResult struct {
	RedirectURL string
	State       string
	MaxAge      int
}

type StartOAuthLinkInput struct {
	Provider  string
	ReturnURL string
	UserID    string
}

type CompleteOAuthInput struct {
	Provider     string
	Code         string
	State        string
	BrowserState string
	UserAgent    string
}

type CompleteOAuthRegistrationInput struct {
	PendingToken string
	Username     string
	UserAgent    string
}

type AddPasswordInput struct {
	UserID   string
	Password string
}

type MeInput struct {
	UserID string
}
