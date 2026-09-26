package auth

type LoginInput struct {
	Login    string
	Password string
}

type LoginResult struct {
	User             User
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
}
