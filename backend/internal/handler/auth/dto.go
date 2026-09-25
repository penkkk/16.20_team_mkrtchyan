package auth

type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
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

type LoginResponse struct {
	AccessToken      string       `json:"accessToken"`
	TokenType        string       `json:"tokenType"`
	ExpiresIn        int          `json:"expiresIn"`
	RefreshExpiresIn int          `json:"refreshExpiresIn"`
	User             UserResponse `json:"user"`
}

type UserResponse struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	TgUsername string `json:"tg_username"`
	Name       string `json:"name"`
	Surname    string `json:"surname"`
}
