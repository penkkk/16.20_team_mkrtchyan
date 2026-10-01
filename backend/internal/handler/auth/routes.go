package auth

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")

	auth.POST("/login", h.login)
	auth.DELETE("/logout", h.logout)
	auth.POST("/register", h.register)
	auth.GET("/refresh", h.refresh)
	auth.GET("/:provider/start", h.startOAuth)
	auth.GET("/:provider/callback", h.oauthCallback)
}
