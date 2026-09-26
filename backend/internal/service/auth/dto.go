package auth

type LoginInput struct {
	Login     string
	Password  string
	UserAgent string
}

type LoginResult struct {
	AccessToken      string
	RefreshToken     string
	TokenType        string
	ExpiresIn        int
	RefreshExpiresIn int
	User             User
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