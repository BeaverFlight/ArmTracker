package handlers

import (
	"auth_service/internal/models"
	"context"
	"crypto/rsa"
	"fmt"
	"log/slog"
	"net/http"
	"pkg/respond"
	"pkg/roles"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthService interface {
	GetToken(ctx context.Context, guid uuid.UUID, role roles.Role) (string, string, error)
	DeleteRefresh(ctx context.Context, refresh string) error
	RefreshTokens(ctx context.Context, access, refresh string) (string, string, error)
	PublicKey() *rsa.PublicKey
}

type Handlers struct {
	srv                 AuthService
	refreshCookieMaxAge int
	log *slog.Logger
}

func NewHandlers(srv AuthService, refreshCookieMaxAge int, log *slog.Logger) *Handlers {
	return &Handlers{srv: srv, refreshCookieMaxAge: refreshCookieMaxAge, log: log}
}

func errorHandler(c *gin.Context, err error) {
	code := models.HTTPCode(err)
	if code == 0 {
		respond.InternalError(c)
		return
	}
	respond.JSON(c, code, err.Error())
}

func (h *Handlers) Authorization(c *gin.Context) {
	req := models.RequestAuthorization{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		respond.BadRequest(c, "Невалидный JSON")
		return
	}

	ctx := c.Request.Context()
	access, refresh, err := h.srv.GetToken(ctx, req.GUID, req.Role)
	if err != nil {
		errorHandler(c, err)
		return
	}

	c.SetCookie(
		"refresh",
		refresh,
		h.refreshCookieMaxAge,
		"/auth/refresh",
		"",
		true,
		true,
	)

	c.JSON(http.StatusOK, gin.H{"access": access})

}

func (h *Handlers) Logout(c *gin.Context) {
	refresh, err := c.Cookie("refresh")

	if err != nil {
		respond.BadRequest(c, "Не найден refresh токен")
		return
	}

	ctx := c.Request.Context()

	err = h.srv.DeleteRefresh(ctx, refresh)
	if err != nil {
		errorHandler(c, err)
		return
	}

	c.SetCookie(
		"refresh",
		"",
		-1,
		"/auth/refresh",
		"",
		true,
		true,
	)

	c.JSON(http.StatusOK, gin.H{"message": "Ok"})
}

func (h *Handlers) Refresh(c *gin.Context) {
	refresh, err := c.Cookie("refresh")
	if err != nil {
		respond.BadRequest(c, "Не найден refresh токен")
		return
	}

	access, err := h.findAccess(c)
	if err != nil {
		respond.BadRequest(c, err.Error())
		return
	}
	ctx := c.Request.Context()

	access, refresh, err = h.srv.RefreshTokens(ctx, access, refresh)
	if err != nil {
		errorHandler(c, err)
		return
	}

	c.SetCookie(
		"refresh",
		refresh,
		h.refreshCookieMaxAge,
		"/auth/refresh",
		"",
		true,
		true,
	)

	c.JSON(http.StatusOK, gin.H{"access": access})
}

func (h *Handlers) findAccess(c *gin.Context) (string, error) {

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("не найден заголовок Authorization")
	}

	access, ok := strings.CutPrefix(authHeader, "Bearer ")
	if !ok || access == "" {
		return "", fmt.Errorf("неаерный формат Authorization заголовка")
	}

	return access, nil

}
