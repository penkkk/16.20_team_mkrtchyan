package auth

type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AddPasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username   string `json:"username" binding:"required"`
	TgUsername string `json:"tg_username"`
	Email      string `json:"email" binding:"required,email"`
	Name       string `json:"name" binding:"required"`
	Surname    string `json:"surname" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

type CompleteOAuthRegistrationRequest struct {
	Username string `json:"username" binding:"required"`
}

type LoginResponse struct {
	User             UserResponse `json:"user"`
	AccessToken      string       `json:"accessToken"`
	TokenType        string       `json:"tokenType"`
	ExpiresIn        int          `json:"expiresIn"`
	RefreshExpiresIn int          `json:"refreshExpiresIn"`
}

type ExternalLinkResponse struct {
	RedirectURL string `json:"redirect_url"`
}

type RefreshResponse struct {
	AccessToken      string `json:"accessToken"`
	TokenType        string `json:"tokenType"`
	ExpiresIn        int    `json:"expiresIn"`
	RefreshExpiresIn int    `json:"refreshExpiresIn"`
}

type UserResponse struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	TgUsername string `json:"tg_username"`
	Name       string `json:"name"`
	Surname    string `json:"surname"`
}
