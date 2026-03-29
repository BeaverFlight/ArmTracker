package service_test

import (
	"context"
	"strings"
	"testing"
	"user_service/internal/models"
	"user_service/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	validLogin := "valid_login"
	validPassword := "valid_password"
	guid := uuid.New()

	tests := []struct {
		name     string
		user     models.User
		repoFn   func(ctx context.Context, user models.User) (uuid.UUID, error)
		wantGUID uuid.UUID
		wantErr  error
	}{
		{
			name: "успешное создание",
			user: models.User{
				Login:    validLogin,
				Password: validPassword,
			},
			repoFn: func(ctx context.Context, user models.User) (uuid.UUID, error) {
				return guid, nil
			},
			wantGUID: guid,
			wantErr:  nil,
		},
		{
			name: "короткий пароль",
			user: models.User{
				Login:    validLogin,
				Password: "short",
			},
			wantErr: models.ErrPasswordIsShort,
		},
		{
			name: "короткий логин",
			user: models.User{
				Login:    "ab",
				Password: validPassword,
			},
			wantErr: models.ErrLoginIsShort,
		},
		{
			name: "длинный пароль",
			user: models.User{
				Login:    validLogin,
				Password: strings.Repeat("long_password", 20),
			},
			wantErr: models.ErrPasswordIsLong,
		},
		{
			name: "длинный логин",
			user: models.User{
				Login:    strings.Repeat("long_login", 20),
				Password: validPassword,
			},
			wantErr: models.ErrLoginIsLong,
		},
		{
			name: "логин занят",
			user: models.User{
				Login:    validLogin,
				Password: validPassword,
			},
			repoFn: func(ctx context.Context, user models.User) (uuid.UUID, error) {
				return uuid.Nil, models.ErrLoginBusy
			},
			wantErr: models.ErrLoginBusy,
		},
		{
			name: "неизвестная ошибка",
			user: models.User{
				Login:    validLogin,
				Password: validPassword,
			},
			repoFn: func(ctx context.Context, user models.User) (uuid.UUID, error) {
				return uuid.Nil, models.ErrUnknown
			},
			wantErr: models.ErrRegistrationFailed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &mockUserRepository{
				CreateUserFn: test.repoFn,
			}

			s := service.NewUserService(repo)

			gotGUID, err := s.CreateUser(context.Background(), test.user)

			assert.ErrorIs(t, err, test.wantErr) // ← исправлено: (got, want)

			assert.Equal(t, test.wantGUID, gotGUID)
		})
	}
}

func TestGetUserByGUID(t *testing.T) {
	validGUID := uuid.New()
	validLogin := "valid_login"

	tests := []struct {
		name      string
		user      models.User
		repoFn    func(ctx context.Context, guid uuid.UUID) (models.User, error)
		wantErr   error
		wantLogin string
	}{
		{
			name: "успешное получение login",
			user: models.User{
				GUID: validGUID,
			},
			repoFn: func(ctx context.Context, guid uuid.UUID) (models.User, error) {
				return models.User{
					Login: validLogin,
				}, nil
			},
			wantLogin: validLogin,
		},
		{
			name: "нулевой guid",
			user: models.User{
				GUID: uuid.Nil,
			},
			wantErr: models.ErrGUIDNotFound,
		},
		{
			name: "не найден пользователь",
			user: models.User{
				GUID: uuid.New(),
			},
			repoFn: func(ctx context.Context, guid uuid.UUID) (models.User, error) {
				return models.User{}, models.ErrGUIDNotFound
			},
			wantErr: models.ErrGUIDNotFound,
		},
		{
			name: "неизвестная ошибка",
			user: models.User{
				GUID: uuid.New(),
			},
			repoFn: func(ctx context.Context, guid uuid.UUID) (models.User, error) {
				return models.User{}, models.ErrUnknown
			},
			wantErr: models.ErrUnknown,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &mockUserRepository{
				GetUserByGUIDFn: test.repoFn,
			}

			s := service.NewUserService(repo)

			user, err := s.GetUserByGUID(context.Background(), test.user.GUID)

			assert.ErrorIs(t, err, test.wantErr)

			assert.Equal(t, test.wantLogin, user.Login)
		})
	}
}

func TestGetUserByLogin(t *testing.T) {
	validGUID := uuid.New()
	validLogin := "valid_login"

	tests := []struct {
		name     string
		user     models.User
		repoFn   func(ctx context.Context, login string) (models.User, error)
		wantGUID uuid.UUID
		wantErr  error
	}{
		{
			name: "успешное получение guid",
			user: models.User{
				Login: validLogin,
			},
			repoFn: func(ctx context.Context, login string) (user models.User, err error) {
				return models.User{
					GUID: validGUID,
				}, nil
			},
			wantGUID: validGUID,
		},
		{
			name: "не найден пользователь",
			user: models.User{
				Login: validLogin,
			},
			repoFn: func(ctx context.Context, login string) (user models.User, err error) {
				return models.User{}, models.ErrLoginNotFound
			},
			wantErr: models.ErrLoginNotFound,
		},
		{
			name: "неизвестная ошибка",
			user: models.User{
				Login: validLogin,
			},
			repoFn: func(ctx context.Context, login string) (user models.User, err error) {
				return models.User{}, models.ErrUnknown
			},
			wantErr: models.ErrUnknown,
		},
		{
			name: "короткий логин",
			user: models.User{
				Login: "123",
			},
			wantErr: models.ErrLoginIsShort,
		},
		{
			name: "длинный логин",
			user: models.User{
				Login: strings.Repeat("a", 256),
			},
			wantErr: models.ErrLoginIsLong,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &mockUserRepository{
				GetUserByLoginFn: test.repoFn,
			}

			s := service.NewUserService(repo)

			user, err := s.GetUserByLogin(context.Background(), test.user.Login)

			assert.ErrorIs(t, err, test.wantErr)

			assert.Equal(t, test.wantGUID, user.GUID)
		})
	}
}

func TestVerifyUser(t *testing.T) {
	validLogin := "valid_login"
	validPassword := "valid_password"
	validGUID := uuid.New()
	tests := []struct {
		name     string
		user     models.User
		repoFn   func(ctx context.Context, login, password string) (uuid.UUID, error)
		wantErr  error
		wantGUID uuid.UUID
	}{
		{
			name: "успешная верификация",
			user: models.User{
				Login:    validLogin,
				Password: validPassword,
			},
			repoFn: func(ctx context.Context, login, password string) (uuid.UUID, error) {
				return validGUID, nil
			},
			wantGUID: validGUID,
		},
		{
			name: "неверный логин",
			user: models.User{
				Login:    "invalid_login",
				Password: validPassword,
			},
			repoFn: func(ctx context.Context, login, password string) (uuid.UUID, error) {
				return uuid.Nil, models.ErrLoginNotFound
			},
			wantErr: models.ErrAuthenticationFailed,
		},
		{
			name: "неверный пароль",
			user: models.User{
				Login:    validLogin,
				Password: "invalid_password",
			},
			repoFn: func(ctx context.Context, login, password string) (uuid.UUID, error) {
				return uuid.Nil, models.ErrInvalidPassword
			},
			wantErr: models.ErrAuthenticationFailed,
		},

		{
			name: "неизвестная ошибка",
			user: models.User{
				Login:    validLogin,
				Password: validPassword,
			},
			repoFn: func(ctx context.Context, login, password string) (uuid.UUID, error) {
				return uuid.Nil, models.ErrUnknown
			},
			wantErr: models.ErrUnknown,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &mockUserRepository{
				VerifyUserFn: test.repoFn,
			}

			s := service.NewUserService(repo)

			guid, err := s.VerifyUser(context.Background(), test.user.Login, test.user.Password)

			assert.ErrorIs(t, err, test.wantErr)

			assert.Equal(t, test.wantGUID, guid)
		})
	}
}

func TestMakeAdmin(t *testing.T) {

	validGUID := uuid.New()
	tests := []struct {
		name    string
		guid    uuid.UUID
		repoFn  func(ctx context.Context, guid uuid.UUID) error
		wantErr error
	}{
		{
			name: "успешное создание",
			guid: validGUID,
			repoFn: func(ctx context.Context, guid uuid.UUID) error {
				return nil
			},
		},
		{
			name:    "нулевой guid",
			guid:    uuid.Nil,
			wantErr: models.ErrGUIDNotFound,
		},
		{
			name: "guid не найден",
			guid: uuid.New(),
			repoFn: func(ctx context.Context, guid uuid.UUID) error {
				return models.ErrGUIDNotFound
			},
			wantErr: models.ErrGUIDNotFound,
		},
		{
			name: "неизвестная ошибка",
			guid: uuid.New(),
			repoFn: func(ctx context.Context, guid uuid.UUID) error {
				return models.ErrUnknown
			},
			wantErr: models.ErrUnknown,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &mockUserRepository{
				MakeAdminFn: test.repoFn,
			}

			s := service.NewUserService(repo)

			err := s.MakeAdmin(context.Background(), test.guid)

			assert.ErrorIs(t, err, test.wantErr)
		})
	}
}

func TestUpdateUser(t *testing.T) {

	validUser := models.User{
		Login:    "valid_login",
		Password: "valid_password",
	}
	tests := []struct {
		name    string
		user    models.User
		repoFn  func(ctx context.Context, user models.User) error
		wantErr error
	}{
		{
			name: "успешное обновление",
			user: validUser,
			repoFn: func(ctx context.Context, user models.User) error {
				return nil
			},
		},
		{
			name: "guid не найден",
			user: validUser,
			repoFn: func(ctx context.Context, user models.User) error {
				return models.ErrGUIDNotFound
			},
			wantErr: models.ErrGUIDNotFound,
		},
		{
			name: "неизвестная ошибка",
			user: validUser,
			repoFn: func(ctx context.Context, user models.User) error {
				return models.ErrUnknown
			},
			wantErr: models.ErrUnknown,
		},
		{
			name: "нет полей для обновления",
			user: validUser,
			repoFn: func(ctx context.Context, user models.User) error {
				return models.ErrNoUpdateFields
			},
			wantErr: models.ErrNoUpdateFields,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &mockUserRepository{
				UpdateUserFn: test.repoFn,
			}

			s := service.NewUserService(repo)

			err := s.UpdateUser(context.Background(), test.user)

			assert.ErrorIs(t, err, test.wantErr)
		})
	}
}

func TestChangePassword(t *testing.T) {
	validPassword := "password"

	tests := []struct {
		name        string
		guid        uuid.UUID
		newPassword string
		oldPassword string
		repoFn      func(ctx context.Context, guid uuid.UUID, oldPassword, newPassword string) error
		wantErr     error
	}{
		{
			name:        "успешная смена пароля",
			guid:        uuid.New(),
			newPassword: validPassword,
			oldPassword: validPassword,
			repoFn: func(ctx context.Context, guid uuid.UUID, oldPassword, newPassword string) error {
				return nil
			},
		},

		{
			name:        "guid не найден",
			guid:        uuid.New(),
			newPassword: validPassword,
			oldPassword: validPassword,
			repoFn: func(ctx context.Context, guid uuid.UUID, oldPassword, newPassword string) error {
				return models.ErrGUIDNotFound
			},
			wantErr: models.ErrGUIDNotFound,
		},

		{
			name:        "неверный старый пароль",
			guid:        uuid.New(),
			newPassword: validPassword,
			oldPassword: "wrong",
			repoFn: func(ctx context.Context, guid uuid.UUID, oldPassword, newPassword string) error {
				return models.ErrInvalidPassword
			},
			wantErr: models.ErrInvalidPassword,
		},
		{
			name:        "неизвестная ошибка",
			guid:        uuid.New(),
			newPassword: validPassword,
			oldPassword: validPassword,
			repoFn: func(ctx context.Context, guid uuid.UUID, oldPassword, newPassword string) error {
				return models.ErrUnknown
			},
			wantErr: models.ErrUnknown,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &mockUserRepository{
				ChangePasswordFn: test.repoFn,
			}

			s := service.NewUserService(repo)

			err := s.ChangePassword(context.Background(), test.guid, test.oldPassword, test.newPassword)

			assert.ErrorIs(t, err, test.wantErr)
		})
	}
}
