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

type CompleteOAuthInput struct {
	Provider     string
	Code         string
	State        string
	BrowserState string
	UserAgent    string
}
