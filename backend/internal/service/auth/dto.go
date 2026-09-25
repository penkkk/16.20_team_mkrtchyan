package auth

type LoginInput struct {
	Login    string
	Password string
}

type LoginResult struct {
	AccessToken      string
	RefreshToken     string
	TokenType        string
	ExpiresIn        int
	RefreshExpiresIn int
	User             User
}

type RegisterInput struct {
	Username   string
	Email      string
	Password   string
	TgUsername *string
	Name       string
	Surname    string
}
