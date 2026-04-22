package main

import (
	"auth_service/internal/handlers"
	"auth_service/internal/repository"
	"auth_service/internal/service"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Генерация RSA ключа при старте
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		logger.Error("генерация RSA ключа", "err", err)
		os.Exit(1)
	}
	logger.Info("RSA-2048 ключ успешно сгенерирован")

	// Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	redisCfg, err := repository.LoadRedisConfig()
	if err != nil {
		logger.Error("загрузка конфига Redis", "err", err)
		os.Exit(1)
	}

	repo, err := repository.NewClient(ctx, redisCfg)
	if err != nil {
		logger.Error("подключение к Redis", "err", err)
		os.Exit(1)
	}
	defer repo.Close()

	// TTL из env
	accessTTL, _ := time.ParseDuration(getEnv("ACCESS_TTL", "15m"))
	refreshTTL, _ := time.ParseDuration(getEnv("REFRESH_TTL", "168h"))
	cookieMaxAge, _ := strconv.Atoi(getEnv("COOKIE_MAX_AGE", "604800"))

	srv := service.NewService(repo, privateKey, accessTTL, refreshTTL)
	handler := handlers.NewHandlers(srv, cookieMaxAge)

	r := setupRouter(handler)

	r.Run()

}

func setupRouter(handler *handlers.Handlers) *gin.Engine {
	r := gin.Default()

	auth := r.Group("/auth")
	{
		auth.POST("/token", handler.Authorization)
		auth.POST("/refresh", handler.Refresh)
		auth.POST("/logout", handler.Logout)
	}

	// Эндпоинты для получения публичного ключа — gateway и другие сервисы
	// обращаются сюда при старте, чтобы кэшировать ключ для валидации JWT
	r.GET("/.well-known/jwks.json", handler.GetPublicKey)
	r.GET("/.well-known/public-key.pem", handler.GetPublicKeyPEM)

	return r
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
