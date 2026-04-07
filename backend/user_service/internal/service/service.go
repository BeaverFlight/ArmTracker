package service

import (
	"context"
	"errors"
	"pkg/roles"
	"user_service/internal/models"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type UserRepository interface {
	GetUserByGUID(ctx context.Context, guid uuid.UUID) (models.User, error)
	GetUserByLogin(ctx context.Context, login string) (models.User, error)
	CreateUser(ctx context.Context, user models.User) (uuid.UUID, error)
	VerifyUser(ctx context.Context, login, password string) (uuid.UUID, roles.Role, error)
	SetRole(ctx context.Context, guid uuid.UUID, role roles.Role) error
	UpdateUser(ctx context.Context, user models.User) error
	ChangePassword(ctx context.Context, guid uuid.UUID, oldPassword, newPassword string) error
}

type Service struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(ctx context.Context, user models.User) (uuid.UUID, error) {

	err := s.validateLogin(user.Login)
	if err != nil {
		return uuid.Nil, err
	}

	err = s.validatePassword(user.Password)
	if err != nil {
		return uuid.Nil, err
	}
	err = validator.New().Var(user.Email, "email")
	if err != nil {
		return uuid.Nil, models.ErrInvalidEmail
	}

	guid, err := s.repo.CreateUser(ctx, user)
	if errors.Is(err, models.ErrLoginBusy) {
		return uuid.Nil, models.ErrLoginBusy
	}

	if err != nil {
		return uuid.Nil, models.ErrRegistrationFailed
	}

	return guid, err
}

func (s *Service) validateLogin(login string) error {
	if len(login) <= 4 {
		return models.ErrLoginIsShort
	}

	if len(login) >= 128 {
		return models.ErrLoginIsLong
	}

	return nil
}

func (s *Service) validatePassword(password string) error {
	if len(password) <= 6 {
		return models.ErrPasswordIsShort
	}

	if len(password) >= 128 {
		return models.ErrPasswordIsLong
	}

	return nil
}

func (s *Service) GetUserByGUID(ctx context.Context, guid uuid.UUID) (models.User, error) {
	if guid == uuid.Nil {
		return models.User{}, models.ErrGUIDNotFound
	}

	user, err := s.repo.GetUserByGUID(ctx, guid)
	if errors.Is(err, models.ErrGUIDNotFound) {
		return models.User{}, models.ErrGUIDNotFound
	}

	if err != nil {
		return models.User{}, models.ErrUnknown
	}

	return user, nil
}

func (s *Service) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	err := s.validateLogin(login)
	if err != nil {
		return models.User{}, err
	}

	user, err := s.repo.GetUserByLogin(ctx, login)

	if errors.Is(err, models.ErrLoginNotFound) {
		return models.User{}, models.ErrLoginNotFound
	}

	if err != nil {
		return models.User{}, models.ErrUnknown
	}

	return user, nil
}

func (s *Service) VerifyUser(ctx context.Context, login, password string) (uuid.UUID, roles.Role, error) {
	err := s.validateLogin(login)
	if err != nil {
		return uuid.Nil, roles.RoleNone, err
	}

	err = s.validatePassword(password)
	if err != nil {
		return uuid.Nil, roles.RoleNone, err
	}

	guid, role, err := s.repo.VerifyUser(ctx, login, password)

	if errors.Is(err, models.ErrInvalidPassword) {
		return uuid.Nil, roles.RoleNone, models.ErrAuthenticationFailed
	}

	if errors.Is(err, models.ErrLoginNotFound) {
		return uuid.Nil, roles.RoleNone, models.ErrAuthenticationFailed
	}

	if err != nil {
		return uuid.Nil, roles.RoleNone, models.ErrUnknown
	}

	return guid, role, nil
}

func (s *Service) SetRole(ctx context.Context, guid uuid.UUID, role roles.Role) error {

	if guid == uuid.Nil {
		return models.ErrGUIDNotFound
	}

	err := s.repo.SetRole(ctx, guid, role)

	if errors.Is(err, models.ErrGUIDNotFound) {
		return models.ErrGUIDNotFound
	}

	if err != nil {
		return models.ErrUnknown
	}

	return nil
}

func (s *Service) UpdateUser(ctx context.Context, user models.User) error {

	err := s.repo.UpdateUser(ctx, user)

	if errors.Is(err, models.ErrGUIDNotFound) {
		return models.ErrGUIDNotFound
	}

	if errors.Is(err, models.ErrNoUpdateFields) {
		return models.ErrNoUpdateFields
	}

	if err != nil {
		return models.ErrUnknown
	}

	return nil
}

func (s *Service) ChangePassword(ctx context.Context, guid uuid.UUID, oldPassword, newPassword string) error {
	err := s.validatePassword(newPassword)
	if err != nil {
		return err
	}

	err = s.repo.ChangePassword(ctx, guid, oldPassword, newPassword)

	if errors.Is(err, models.ErrGUIDNotFound) {
		return models.ErrGUIDNotFound
	}

	if errors.Is(err, models.ErrInvalidPassword) {
		// TODO: Очень жёстко следить за сменой пароля
		return models.ErrInvalidPassword
	}

	if err != nil {
		return models.ErrUnknown
	}

	return nil
}
