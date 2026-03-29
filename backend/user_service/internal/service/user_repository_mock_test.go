package service_test

import (
	"context"
	"user_service/internal/models"

	"github.com/google/uuid"
)

type mockUserRepository struct {
	GetUserByGUIDFn  func(ctx context.Context, guid uuid.UUID) (models.User, error)
	GetUserByLoginFn func(ctx context.Context, login string) (models.User, error)
	CreateUserFn     func(ctx context.Context, user models.User) (uuid.UUID, error)
	VerifyUserFn     func(ctx context.Context, login, password string) (uuid.UUID, error)
	MakeAdminFn      func(ctx context.Context, guid uuid.UUID) error
	UpdateUserFn     func(ctx context.Context, user models.User) error
	ChangePasswordFn func(ctx context.Context, guid uuid.UUID, oldPassword, newPassword string) error
}

func (m *mockUserRepository) GetUserByGUID(ctx context.Context, guid uuid.UUID) (models.User, error) {
	return m.GetUserByGUIDFn(ctx, guid)
}

func (m *mockUserRepository) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	return m.GetUserByLoginFn(ctx, login)
}

func (m *mockUserRepository) CreateUser(ctx context.Context, user models.User) (uuid.UUID, error) {
	return m.CreateUserFn(ctx, user)
}

func (m *mockUserRepository) VerifyUser(ctx context.Context, login, password string) (uuid.UUID, error) {
	return m.VerifyUserFn(ctx, login, password)
}

func (m *mockUserRepository) MakeAdmin(ctx context.Context, guid uuid.UUID) error {
	return m.MakeAdminFn(ctx, guid)
}

func (m *mockUserRepository) UpdateUser(ctx context.Context, user models.User) error {
	return m.UpdateUserFn(ctx, user)
}

func (m *mockUserRepository) ChangePassword(ctx context.Context, guid uuid.UUID, oldPassword, newPassword string) error {
	return m.ChangePasswordFn(ctx, guid, oldPassword, newPassword)
}
