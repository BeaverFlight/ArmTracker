package main

import (
	"auth_service/internal/handlers"
	"auth_service/internal/repository"
	"auth_service/internal/service"
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	repo := repository.NewClient(ctx)
	defer repo.Close()

	srv := service.NewService(repo)

	handler := handlers.NewHandlers(srv)

	r := setupRouter(*handler)

	err = r.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func setupRouter(handler handlers.Handlers) *gin.Engine {
	r := gin.Default()

	return r
}
