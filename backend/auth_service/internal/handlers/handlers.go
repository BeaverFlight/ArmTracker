package handlers

import (
	"auth_service/internal/models"
	"context"
	"net/http"
	"pkg/respond"
	"pkg/roles"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthService interface {
	GetToken(ctx context.Context, guid uuid.UUID, role roles.Role) (string, string, error)
	DeleteRefresh(ctx context.Context, refresh string) error
	RefreshTokens(ctx context.Context, refresh string) (string, string, error)
}

type Handlers struct {
	srv                 AuthService
	refreshCookieMaxAge int
}

func NewHandlers(srv AuthService, refreshCookieMaxAge int) *Handlers {
	return &Handlers{srv: srv, refreshCookieMaxAge: refreshCookieMaxAge}
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
		h.refreshCookieMaxAge,
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

	ctx := c.Request.Context()

	access, refresh, err := h.srv.RefreshTokens(ctx, refresh)
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
