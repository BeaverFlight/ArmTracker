package main

import (
	"context"
	"log"
	"pkg/psql"
	"time"
	"user_service/internal/database"
	"user_service/internal/handlers"
	"user_service/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := psql.LoadPSQLConfig()
	if err != nil {
		log.Panic(err)
	}
	pool, err := psql.NewPSQL(ctx, *cfg)

	repo := database.NewDatabase(pool)
	defer repo.Close()

	srv := service.NewUserService(repo)

	handler := handlers.NewHandlers(srv)

	r.POST("/user", handler.CreateUser)
	r.POST("/user/auth", handler.VerifyUser)

	r.PATCH("/user", handler.UpdateUser)

	r.PUT("/user/:guid/:role", handler.SetRole)
	r.PUT("/user/password", handler.ChangePassword)

	r.GET("/user/:guid", handler.GetUserByGUID)
	r.GET("/user", handler.GetUserByLogin) // query
	r.Run()
}
