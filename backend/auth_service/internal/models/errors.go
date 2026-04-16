package models

import (
	"errors"
	"net/http"
)

var (
	ErrUnknown             = errors.New("неизвестная ошибка")
	ErrAccessCreation      = errors.New("ошибка создания access токена")
	ErrInvalidAccessToken  = errors.New("access token не валиден")
	ErrInvalidRefreshToken = errors.New("refresh token не валиден")
	ErrTokenMismatch       = errors.New("несоответствие токенов")
	ErrTokenNotFound       = errors.New("токены не найдены")
)

var knownErrors = []struct {
	err  error
	code int
}{
	{ErrAccessCreation, http.StatusInternalServerError},
	{ErrInvalidAccessToken, http.StatusBadRequest},
	{ErrInvalidRefreshToken, http.StatusBadRequest},
	{ErrTokenNotFound, http.StatusBadRequest},
	{ErrTokenMismatch, http.StatusBadRequest},
}

func HTTPCode(err error) int {
	for _, e := range knownErrors {
		if errors.Is(err, e.err) {
			return e.code
		}
	}
	return 0
}
