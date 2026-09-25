package auth

import authservice "opd/internal/service/auth"

type Handler struct {
	service authservice.Service
}

func NewHandler(service authservice.Service) *Handler {
	return &Handler{
		service: service,
	}
}
