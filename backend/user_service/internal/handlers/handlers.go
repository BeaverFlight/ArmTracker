package handlers

import (
	"context"
	"net/http"
	"user_service/internal/models"

	"pkg/respond"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserService interface {
	CreateUser(ctx context.Context, user models.User) (uuid.UUID, error)
	GetUserByGUID(ctx context.Context, guid uuid.UUID) (models.User, error)
	GetUserByLogin(ctx context.Context, login string) (models.User, error)
	VerifyUser(ctx context.Context, login, password string) (uuid.UUID, error)
	MakeAdmin(ctx context.Context, guid uuid.UUID) error
	UpdateUser(ctx context.Context, user models.User) error
	ChangePassword(ctx context.Context, guid uuid.UUID, oldPassword, newPassword string) error
}

type Handlers struct {
	srv UserService
}

func errorHandler(c *gin.Context, err error) {
	code := models.HTTPCode(err)
	if code == 0 {
		respond.InternalError(c)
		return
	}
	respond.JSON(c, code, err.Error())
}
func NewHandlers(srv UserService) *Handlers {
	return &Handlers{srv: srv}
}

func (h *Handlers) CreateUser(c *gin.Context) {
	user := models.User{}
	if err := c.ShouldBindBodyWithJSON(&user); err != nil {
		respond.BadRequest(c, "Невалидный JSON")
		return
	}

	ctx := c.Request.Context()
	guid, err := h.srv.CreateUser(ctx, user)

	if err != nil {
		errorHandler(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"guid": guid})

}

func (h *Handlers) GetUserByGUID(c *gin.Context) {
	guid, err := uuid.Parse(c.Param("guid"))
	if err != nil {
		respond.BadRequest(c, "Невалидный GUID")
		return
	}

	ctx := c.Request.Context()
	user, err := h.srv.GetUserByGUID(ctx, guid)
	if err != nil {
		errorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handlers) GetUserByLogin(c *gin.Context) {
	login := c.Query("login")

	ctx := c.Request.Context()
	user, err := h.srv.GetUserByLogin(ctx, login)
	if err != nil {
		errorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handlers) VerifyUser(c *gin.Context) {
	user := models.User{}
	if err := c.ShouldBindBodyWithJSON(&user); err != nil {
		respond.BadRequest(c, "Невалидный JSON")
		return
	}

	ctx := c.Request.Context()
	guid, err := h.srv.VerifyUser(ctx, user.Login, user.Password)
	if err != nil {
		errorHandler(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"guid": guid})
}

func (h *Handlers) MakeAdmin(c *gin.Context) {
	guid, err := uuid.Parse(c.Param("guid"))
	if err != nil {
		respond.BadRequest(c, "Невалидный GUID")
		return
	}

	ctx := c.Request.Context()
	err = h.srv.MakeAdmin(ctx, guid)
	if err != nil {
		errorHandler(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"guid": guid})
}

func (h *Handlers) UpdateUser(c *gin.Context) {
	user := models.User{}
	if err := c.ShouldBindBodyWithJSON(&user); err != nil {
		respond.BadRequest(c, "Невалидный JSON")
		return
	}

	ctx := c.Request.Context()
	err := h.srv.UpdateUser(ctx, user)
	if err != nil {
		errorHandler(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"guid": user.GUID})
}

func (h *Handlers) ChangePassword(c *gin.Context) {
	password := models.UserChangePassword{}
	if err := c.ShouldBindBodyWithJSON(&password); err != nil {
		respond.BadRequest(c, "Невалидный JSON")
		return
	}

	ctx := c.Request.Context()
	err := h.srv.ChangePassword(ctx, password.GUID, password.OldPassword, password.NewPassword)
	if err != nil {
		errorHandler(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"guid": password.GUID})
}
