package models

import (
	"errors"
	"net/http"
)

// Ошибки бизнес-логики
var (
	ErrLoginIsShort         = errors.New("логин слишком короткий")
	ErrLoginIsLong          = errors.New("логин слишком длинный")
	ErrPasswordIsShort      = errors.New("пароль слишком короткий")
	ErrPasswordIsLong       = errors.New("пароль слишком длинный")
	ErrNameIsLong           = errors.New("имя слишком длинное")
	ErrRegistrationFailed   = errors.New("регистрация не удалась")
	ErrUnknown              = errors.New("неизвестная ошибка")
	ErrAuthenticationFailed = errors.New("неправильный логин или пароль")
)

// Ошибки базы данных
var (
	ErrLoginBusy       = errors.New("данный логин уже занят")
	ErrLoginNotFound   = errors.New("пользователь ввёл неверный логин")
	ErrInvalidPassword = errors.New("пользователь ввёл неверный пароль")
	ErrGUIDNotFound    = errors.New("пользователь не найден")
	ErrNoUpdateFields  = errors.New("нет полей для обновления")
)

var knownErrors = []struct {
	err  error
	code int
}{
	{ErrLoginBusy, http.StatusConflict},
	{ErrLoginIsShort, http.StatusBadRequest},
	{ErrLoginIsLong, http.StatusBadRequest},
	{ErrPasswordIsShort, http.StatusBadRequest},
	{ErrPasswordIsLong, http.StatusBadRequest},
	{ErrGUIDNotFound, http.StatusNotFound},
	{ErrLoginNotFound, http.StatusNotFound},
	{ErrAuthenticationFailed, http.StatusUnauthorized},
	{ErrNoUpdateFields, http.StatusBadRequest},
	{ErrNameIsLong, http.StatusBadRequest},
}

func HTTPCode(err error) int {
	for _, e := range knownErrors {
		if errors.Is(err, e.err) {
			return e.code
		}
	}
	return 0
}
