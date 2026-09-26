package auth

import authservice "opd/internal/service/auth"

type Handler struct {
	service authservice.AuthService
}

func NewHandler(service authservice.AuthService) *Handler {
	return &Handler{
		service: service,
	}
}
