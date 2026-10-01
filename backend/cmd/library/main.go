package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"opd/internal/config"
	authhandler "opd/internal/handler/auth"
	"opd/internal/handler/middleware"
	"opd/internal/repository"
	"opd/internal/service"
	authservice "opd/internal/service/auth"
	token "opd/internal/service/token"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
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

	err = db.Ping(ctx)
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}

	redisOptions, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("parse redis url: %v", err)
	}

	client := redis.NewClient(redisOptions)

	_, err = client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("connect to redis: %v", err)
	}

	defer client.Close()

	tokenManager := token.NewJWTManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience, 15*time.Minute)
	repositories := repository.NewRepositoriesWithTxManager(db)
	services := service.NewServices(repositories, tokenManager, client, authservice.OAuthConfig{
		GoogleClientID:     cfg.GoogleOAuthClientID,
		YandexClientID:     cfg.YandexOAuthClientID,
		AppPublicURL:       cfg.AppPublicURL,
		YandexClientSecret: cfg.YandexOAuthClientSecret,
		GoogleClientSecret: cfg.GoogleOAuthClientSecret,
	})
	authHandler := authhandler.NewHandler(services.Auth)

	router := gin.Default()

	authRequired := middleware.AuthMiddleware(tokenManager, client)

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		authHandler.RegisterRoutes(v1, authRequired)
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
