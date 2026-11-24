package book

import bookservice "opd/internal/service/book"

type Handler struct {
	service bookservice.BookService
}

func NewHandler(service bookservice.BookService) *Handler {
	return &Handler{service: service}
}
