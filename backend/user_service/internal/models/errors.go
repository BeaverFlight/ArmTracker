package models

import "errors"

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
