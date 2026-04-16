package repository

import (
	"auth_service/internal/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var prefix = "refresh:"

type RedisConfig struct {
	Addr        string        `json:"addr"`
	Password    string        `json:"password"`
	User        string        `json:"user"`
	DB          int           `json:"db"`
	MaxRetries  int           `json:"max_retries"`
	DialTimeout time.Duration `json:"dial_timeout"`
	Timeout     time.Duration `json:"timeout"`
}

type Repository struct {
	rdb *redis.Client
}

func NewClient(ctx context.Context, cfg RedisConfig) (*Repository, error) {
	db := &Repository{rdb: redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		Username:     cfg.User,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	})}
	if err := db.rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ошибка подключения к redis: %w", err)
	}

	return db, nil
}

func (r *Repository) SaveRefresh(ctx context.Context, data models.RefreshData) error {
	payload, err := json.Marshal(data)

	if err != nil {
		return fmt.Errorf("ошибка маршализации RefreshData: %w", err)
	}

	if err := r.rdb.Set(ctx, prefix+data.Refresh.String(), payload, data.RefreshTTl).Err(); err != nil {
		return fmt.Errorf("ошибка redis set: %w", err)
	}

	return nil
}

func (r *Repository) GetRefresh(ctx context.Context, refresh uuid.UUID) (models.RefreshData, error) {
	raw, err := r.rdb.Get(ctx, prefix+refresh.String()).Bytes()

	if errors.Is(err, redis.Nil) {
		return models.RefreshData{}, models.ErrTokenNotFound
	}

	if err != nil {
		return models.RefreshData{}, fmt.Errorf("ошибка redis Get: %w", err)
	}

	data := models.RefreshData{}

	if err := json.Unmarshal(raw, &data); err != nil {
		return models.RefreshData{}, fmt.Errorf("ошибка unmarshal refresh data: %w", err)
	}

	return data, nil
}

func (r *Repository) DeleteRefresh(ctx context.Context, refresh uuid.UUID) error {
	if err := r.rdb.Del(ctx, prefix+refresh.String()).Err(); err != nil {
		return fmt.Errorf("ошибка redis del: %w", err)
	}

	return nil

}
