package main

import (
	"auth_service/internal/handlers"
	"auth_service/internal/repository"
	"auth_service/internal/service"
	"context"
	"os"
	"pkg/logger"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	log := logger.New(getEnv("LOGGER", "local"))

	privateKey, err := service.LoadOrGeneratePrivateKey(getEnv("RSA_KEY_PATH", "./key/private.pem"), log)
	if err != nil {
		log.Error("инициализация RSA ключа", logger.Err(err))
		os.Exit(1)
	}
	log.Info("RSA-2048 ключ успешно сгенерирован")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	redisCfg, err := repository.LoadRedisConfig()
	if err != nil {
		log.Error("загрузка конфига Redis", logger.Err(err))
		os.Exit(1)
	}

	repo, err := repository.NewClient(ctx, redisCfg)
	if err != nil {
		log.Error("подключение к Redis", logger.Err(err))
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
