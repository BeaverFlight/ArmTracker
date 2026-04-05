package handlers_test

import (
	"context"
	"user_service/internal/models"

	"github.com/google/uuid"
)

type mockUserService struct {
	CreateUserFn     func(ctx context.Context, user models.User) (uuid.UUID, error)
	GetUserByGUIDFn  func(ctx context.Context, guid uuid.UUID) (models.User, error)
	GetUserByLoginFn func(ctx context.Context, login string) (models.User, error)
	VerifyUserFn     func(ctx context.Context, login, password string) (uuid.UUID, error)
	MakeAdminFn      func(ctx context.Context, guid uuid.UUID) error
	UpdateUserFn     func(ctx context.Context, user models.User) error
	ChangePasswordFn func(ctx context.Context, guid uuid.UUID, oldPassword, newPassword string) error
}

func (m *mockUserService) CreateUser(ctx context.Context, user models.User) (uuid.UUID, error) {
	return m.CreateUserFn(ctx, user)
}

func (m *mockUserService) GetUserByGUID(ctx context.Context, guid uuid.UUID) (models.User, error) {
	return m.GetUserByGUIDFn(ctx, guid)
}

func (m *mockUserService) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	return m.GetUserByLoginFn(ctx, login)
}

func (m *mockUserService) VerifyUser(ctx context.Context, login, password string) (uuid.UUID, error) {
	return m.VerifyUserFn(ctx, login, password)
}

func (m *mockUserService) MakeAdmin(ctx context.Context, guid uuid.UUID) error {
	return m.MakeAdminFn(ctx, guid)
}

func (m *mockUserService) UpdateUser(ctx context.Context, user models.User) error {
	return m.UpdateUserFn(ctx, user)
}

func (m *mockUserService) ChangePassword(ctx context.Context, guid uuid.UUID, oldPassword, newPassword string) error {
	return m.ChangePasswordFn(ctx, guid, oldPassword, newPassword)
}
