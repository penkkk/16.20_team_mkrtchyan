package book

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterRoutes(router *gin.RouterGroup, authRequired, adminRequired gin.HandlerFunc) {
	admin := router.Group("/admin")

	admin.GET("/books", authRequired, adminRequired, h.list)
	admin.GET("/books/:id/download", authRequired, adminRequired, h.download)
	admin.DELETE("/books/:id", authRequired, adminRequired, h.delete)
	admin.POST("/upload", authRequired, adminRequired, h.upload)
}
