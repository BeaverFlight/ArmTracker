package psql_test

import (
	"context"
	"pkg/psql"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestMain(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	config := LoadConfig(t)
	NewPSQL_Failure(t, ctx)
	NewPSQL_Success(t, ctx, *config)
}

func LoadConfig(t *testing.T) *psql.PSQLConfig {
	host := "local"
	port := 5432
	user := "test_user"
	password := "GylSSS"
	dbName := "MainDB"
	sslMode := "disable"
	maxConns := int32(10)
	minConns := int32(1)
	maxConnIdleTime := 30

	t.Setenv("PSQL_HOST", host)
	t.Setenv("PSQL_PORT", strconv.Itoa(port))
	t.Setenv("PSQL_USER", user)
	t.Setenv("PSQL_PASSWORD", password)
	t.Setenv("PSQL_DBNAME", dbName)
	t.Setenv("PSQL_SSLMODE", sslMode)
	t.Setenv("PSQL_MAXCONNS", strconv.Itoa(int(maxConns)))
	t.Setenv("PSQL_MINCONNS", strconv.Itoa(int(minConns)))
	t.Setenv("PSQL_MAXCONNIDLETIME", strconv.Itoa(maxConnIdleTime))

	config, err := psql.LoadPSQLConfig()
	assert.NoError(t, err)

	assert.Equal(t, host, config.Host)
	assert.Equal(t, port, config.Port)
	assert.Equal(t, user, config.User)
	assert.Equal(t, password, config.Password)
	assert.Equal(t, dbName, config.DBName)
	assert.Equal(t, sslMode, config.SSLMode)
	assert.Equal(t, maxConns, config.MaxConns)
	assert.Equal(t, minConns, config.MinConns)
	assert.Equal(t, time.Duration(maxConnIdleTime*int(time.Second)), config.MaxConnIdleTime)

	return config
}

func NewPSQL_Success(t *testing.T, ctx context.Context, cfg psql.PSQLConfig) {
	pgContainer, err := postgres.Run(
		ctx,
		"postgres:15-alpine",
		postgres.WithDatabase(cfg.DBName),
		postgres.WithUsername(cfg.User),
		postgres.WithPassword(cfg.Password),
		postgres.BasicWaitStrategies(),
	)
	assert.NoError(t, err)

	cfg.Host, err = pgContainer.Host(ctx)
	assert.NoError(t, err)

	port, err := pgContainer.MappedPort(ctx, "5432")
	assert.NoError(t, err)

	cfg.Port = port.Int()

	psql, err := psql.NewPSQL(ctx, cfg)
	assert.NoError(t, err)

	err = psql.Pool.Ping(ctx)
	assert.NoError(t, err)

	psql.Close()
}

func NewPSQL_Failure(t *testing.T, ctx context.Context) {
	_, err := psql.NewPSQL(ctx, psql.PSQLConfig{})
	assert.Error(t, err)
}
