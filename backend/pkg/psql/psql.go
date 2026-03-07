package psql

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type PSQLConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxConns        int32
	MinConns        int32
	MaxConnIdleTime time.Duration
}

type PSQL struct {
	Pool *pgxpool.Pool
}

func LoadPSQLConfig() (*PSQLConfig, error) {
	port, err := strconv.Atoi(os.Getenv("PSQL_PORT"))
	if err != nil {
		return nil, err
	}

	maxConns, err := strconv.Atoi(os.Getenv("PSQL_MAXCONNS"))
	if err != nil {
		return nil, err
	}
	minConns, err := strconv.Atoi(os.Getenv("PSQL_MINCONNS"))
	if err != nil {
		return nil, err
	}
	maxConnIdleTime, err := time.ParseDuration(os.Getenv("PSQL_MAXCONNIDLETIME"))
	if err != nil {
		return nil, err
	}

	return &PSQLConfig{
		Host:            os.Getenv("PSQL_HOST"),
		Port:            port,
		User:            os.Getenv("PSQL_USER"),
		Password:        os.Getenv("PSQL_PASSWORD"),
		DBName:          os.Getenv("PSQL_DBNAME"),
		SSLMode:         os.Getenv("PSQL_SSLMODE"),
		MaxConns:        int32(maxConns),
		MinConns:        int32(minConns),
		MaxConnIdleTime: maxConnIdleTime,
	}, nil
}

func NewPSQL(ctx context.Context, cfg PSQLConfig) (*PSQL, error) {
	poolConfig := &pgxpool.Config{}
	poolConfig.ConnConfig.Host = cfg.Host
	poolConfig.ConnConfig.Port = uint16(cfg.Port)
	poolConfig.ConnConfig.User = cfg.User
	poolConfig.ConnConfig.Password = cfg.Password
	poolConfig.ConnConfig.Database = cfg.DBName
	poolConfig.ConnConfig.TLSConfig = sslModeToTLSConfig(cfg.Host, cfg.SSLMode)

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &PSQL{Pool: pool}, nil
}

func (p *PSQL) Close() {
	p.Pool.Close()
}

func sslModeToTLSConfig(host, mode string) *tls.Config {
	switch mode {
	case "disable":
		return nil
	case "require":
		return &tls.Config{InsecureSkipVerify: true}
	case "verify-full":
		return &tls.Config{ServerName: host}
	default:
		return nil
	}
}

func (p *PSQL) RunMigrations(ctx context.Context, path string) error {
	sqlDB := stdlib.OpenDBFromPool(p.Pool)
	defer sqlDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("Ошибка установки диалекта: %w", err)
	}

	if path == "" {
		path = "internal/dbwork/migrations"
	}

	if err := goose.Up(sqlDB, path); err != nil {
		return fmt.Errorf("Ошибка применения миграций: %w", err)
	}

	return nil
}
