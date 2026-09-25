package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"opd/internal/config"
	authhandler "opd/internal/handler/auth"
	"opd/internal/repository"
	"opd/internal/service"
	token "opd/internal/service/token"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()

	db, err := pgxpool.New(ctx, cfg.DataBaseURL)
	if err != nil {
		log.Fatalf("create db pool: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("connect to db: %v", err)
	}

	tokenManager := token.NewJWTManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience, 15*time.Minute)
	repositories := repository.NewRepositories(db)
	services := service.NewServices(repositories, tokenManager)
	authHandler := authhandler.NewHandler(services.Auth)

	router := gin.Default()

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		authHandler.RegisterRoutes(v1)
	}

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	if err := router.Run(cfg.HTTPPort); err != nil {
		log.Fatalf("run http server: %v", err)
	}
}
