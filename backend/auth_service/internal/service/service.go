package service

import (
	"auth_service/internal/models"
	"context"
	"crypto/rsa"
	"fmt"
	"pkg/roles"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthRepository interface {
	SaveRefresh(ctx context.Context, data models.RefreshData) error
	GetRefreshData(ctx context.Context, refresh uuid.UUID) (models.RefreshData, error)
	DeleteRefreshData(ctx context.Context, refresh uuid.UUID) error
}

type Service struct {
	repo       AuthRepository
	privateKey *rsa.PrivateKey
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type Claims struct {
	GUID   uuid.UUID
	Role   roles.Role
	PairID uuid.UUID
	jwt.RegisteredClaims
}

func NewService(repo AuthRepository, privateKey *rsa.PrivateKey, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{repo: repo, privateKey: privateKey, accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (s *Service) GetToken(ctx context.Context, guid uuid.UUID, role roles.Role) (accessToken, refreshToken string, err error) {
	pairID := uuid.New()
	refresh := uuid.New()

	access, err := s.issueAccessToken(guid, pairID, role)
	if err != nil {
		return "", "", models.ErrAccessCreation
	}

	err = s.repo.SaveRefresh(ctx, models.RefreshData{
		Refresh:    refresh,
		Guid:       guid,
		PairID:     pairID,
		Role:       role,
		RefreshTTL: s.refreshTTL,
	})

	if err != nil {
		return "", "", models.ErrUnknown
	}

	return access, refresh.String(), nil
}

func (s *Service) issueAccessToken(guid, pairID uuid.UUID, role roles.Role) (string, error) {
	now := time.Now()
	claims := Claims{
		GUID:   guid,
		Role:   role,
		PairID: pairID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}

	return signed, nil
}

func (s *Service) RefreshTokens(ctx context.Context, access, refresh string) (access_token, refresh_token string, err error) {
	refreshID, err := uuid.Parse(refresh)
	if err != nil {
		return "", "", models.ErrInvalidRefreshToken
	}

	claims, err := s.parseIgnoreExpiry(access)
	if err != nil {
		return "", "", models.ErrInvalidAccessToken
	}

	data, err := s.repo.GetRefreshData(ctx, refreshID)
	if err != nil {
		return "", "", models.ErrUnknown
	}

	if claims.PairID != data.PairID {
		s.repo.DeleteRefreshData(ctx, refreshID)
		return "", "", models.ErrTokenMismatch
	}

	if err := s.repo.DeleteRefreshData(ctx, refreshID); err != nil {
		return "", "", models.ErrUnknown
	}

	return s.GetToken(ctx, data.Guid, data.Role)
}

func (s *Service) parseIgnoreExpiry(access string) (*Claims, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
		jwt.WithoutClaimsValidation(),
	)

	token, err := parser.ParseWithClaims(access, &Claims{}, func(t *jwt.Token) (any, error) {
		return &s.privateKey.PublicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse access token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, models.ErrInvalidAccessToken
	}

	return claims, err
}

func (s *Service) DeleteRefresh(ctx context.Context, refresh string) error {
	refreshID, err := uuid.Parse(refresh)
	if err != nil {
		return models.ErrInvalidRefreshToken
	}

	err = s.repo.DeleteRefreshData(ctx, refreshID)
	if err != nil {
		return models.ErrUnknown
	}

	return nil
}

func (s *Service) PublicKey() *rsa.PublicKey {
	return &s.privateKey.PublicKey
}
