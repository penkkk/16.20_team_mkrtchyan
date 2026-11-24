package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"opd/internal/config"
	authhandler "opd/internal/handler/auth"
	bookhandler "opd/internal/handler/book"
	"opd/internal/handler/middleware"
	"opd/internal/repository"
	"opd/internal/service"
	authservice "opd/internal/service/auth"
	bookservice "opd/internal/service/book"
	token "opd/internal/service/token"
	"opd/internal/storage"

	_ "opd/docs"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Library API
// @version 1.0
// @description Library backend API.
// @host localhost
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
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
	if cfg.S3BooksBucket == cfg.S3CoversBucket {
		log.Fatal("book and cover buckets must be different")
	}
	if cfg.S3SecretKey == "" {
		log.Fatal("GARAGE_DEFAULT_SECRET_KEY is required")
	}
	var bootstrapErr error
	for attempt := 0; attempt < 30; attempt++ {
		bootstrapErr = storage.BootstrapBuckets(ctx, cfg.S3AdminEndpoint, cfg.S3AdminToken, cfg.S3AccessKey, cfg.S3BooksBucket, cfg.S3CoversBucket)
		if bootstrapErr == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if bootstrapErr != nil {
		log.Fatalf("initialize Garage buckets: %v", bootstrapErr)
	}
	s3store := storage.NewS3(cfg.S3Endpoint, cfg.S3Region, cfg.S3AccessKey, cfg.S3SecretKey)
	bookService := bookservice.NewBookService(repositories.Book, s3store, cfg.S3BooksBucket, cfg.S3CoversBucket)
	bookHandler := bookhandler.NewHandler(bookService)

	router := gin.Default()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authRequired := middleware.AuthMiddleware(tokenManager, client)
	adminRequired := middleware.RequireAdmin(repositories.Auth)

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		authHandler.RegisterRoutes(v1, authRequired)
		bookHandler.RegisterRoutes(v1, authRequired, adminRequired)
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
