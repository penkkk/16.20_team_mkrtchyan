package auth

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterRoutes(router *gin.RouterGroup, authRequired gin.HandlerFunc) {
	auth := router.Group("/auth")

	router.GET("/me", authRequired, h.me)

	auth.POST("/login", h.login)
	auth.DELETE("/logout", h.logout)
	auth.POST("/register", h.register)
	auth.GET("/refresh", h.refresh)
	auth.GET("/:provider/start", h.startOAuth)
	auth.GET("/:provider/callback", h.oauthCallback)
	auth.GET("/:provider/link/start", authRequired, h.linkExternal)
	auth.POST("/password", authRequired, h.addPassword)
}
