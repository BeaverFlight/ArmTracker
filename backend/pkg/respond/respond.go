package respond

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Message string `json:"message"`
}

func JSON(c *gin.Context, code int, message string) {
	c.JSON(code, Response{Message: message})
}

func BadRequest(c *gin.Context, message string) {
	JSON(c, http.StatusBadRequest, message)
}

func InternalError(c *gin.Context) {
	JSON(c, http.StatusInternalServerError, "Внутренняя ошибка сервера")
}

func Conflict(c *gin.Context, message string) {
	JSON(c, http.StatusConflict, message)
}
