package database_test

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"pkg/psql"
	"user_service/internal/database"
	"user_service/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestMain(t *testing.T) {
	db, clean, err := setupTestDB()
	assert.NoError(t, err)

	defer clean()

	CreateUser_Success(t, db)
	CreateUser_LoginBusy(t, db)
	VerifyUser_Success(t, db)
	VerifyUser_InvalidPassword(t, db)
	VerifyUser_InvalidLogin(t, db)
	MakeAdmin_Success(t, db)
	ChangePassword_Success(t, db)
	ChangePassword_InvalidGUID(t, db)
	ChangePassword_WrongOldPassword(t, db)
	UpdateUser_InvalidGUID(t, db)
	UpdateUser_Success(t, db)
	MakeAdmin_InvalidGUID(t, db)

}

func CreateUser_Success(t *testing.T, db *database.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	user := models.User{
		Login:    "Aboba777",
		Password: "#KCV_!#FMOQEIG@#",
		Name:     "Василий Негодник",
		Height:   235,
		Weight:   150,
		Age:      19,
		Admin:    false,
	}
	guid, err := db.CreateUser(ctx, user)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, guid)
}

func CreateUser_LoginBusy(t *testing.T, db *database.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	user := models.User{
		Login:    "BOSS_KFS",
		Password: "#KCV_!#FMOQEIG@#",
		Name:     "Гриша Буборин",
	}
	guid, err := db.CreateUser(ctx, user)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, guid)

	user = models.User{
		Login:    "BOSS_KFS",
		Password: "FEWge23rtg3re",
		Name:     "Били Шмили",
	}
	guid, err = db.CreateUser(ctx, user)
	assert.Equal(t, uuid.Nil, guid)
	assert.Equal(t, models.ErrLoginBusy, err)
}

func VerifyUser_Success(t *testing.T, db *database.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	user := models.User{
		Login:    "Pikmi_Banny",
		Password: "FEWge23rtg3re",
		Name:     "Вася Ржавый",
	}
	var err error

	user.GUID, err = db.CreateUser(ctx, user)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.GUID)

	guid, err := db.VerifyUser(ctx, user.Login, user.Password)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, guid)
	assert.Equal(t, user.GUID, guid)
}

func VerifyUser_InvalidPassword(t *testing.T, db *database.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	user := models.User{
		Login:    "ldfmwepv---",
		Password: "Aefwbewe",
	}
	var err error

	user.GUID, err = db.CreateUser(ctx, user)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.GUID)

	guid, err := db.VerifyUser(ctx, user.Login, "wrong_password")
	assert.Equal(t, uuid.Nil, guid)
	assert.ErrorIs(t, err, models.ErrInvalidPassword, err)
}

func VerifyUser_InvalidLogin(t *testing.T, db *database.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	guid, err := db.VerifyUser(ctx, "wrong_login", "---------")
	assert.Equal(t, uuid.Nil, guid)
	assert.ErrorIs(t, err, models.ErrLoginNotFound, err)
}

func MakeAdmin_Success(t *testing.T, db *database.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	user := models.User{
		Login:    "admin_user",
		Password: "admin_password",
	}
	var err error

	user.GUID, err = db.CreateUser(ctx, user)
	assert.NoError(t, err)

	err = db.MakeAdmin(ctx, user.GUID)
	assert.NoError(t, err)

	user, err = db.GetUserByGUID(ctx, user.GUID)
	assert.NoError(t, err)
	assert.True(t, user.Admin)
}

func MakeAdmin_InvalidGUID(t *testing.T, db *database.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	err := db.MakeAdmin(ctx, uuid.Nil)
	assert.ErrorIs(t, err, models.ErrGUIDNotFound)
}

func ChangePassword_Success(t *testing.T, db *database.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	user := models.User{
		Login:    "test_user",
		Password: "test_password",
	}
	var err error

	user.GUID, err = db.CreateUser(ctx, user)
	assert.NoError(t, err)

	err = db.ChangePassword(ctx, user.GUID, user.Password, "new_password")
	assert.NoError(t, err)
}

func ChangePassword_InvalidGUID(t *testing.T, db *database.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	err := db.ChangePassword(ctx, uuid.Nil, "old_password", "new_password")
	assert.Error(t, err)
}

func ChangePassword_WrongOldPassword(t *testing.T, db *database.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	user := models.User{
		Login:    "test_user_2",
		Password: "test_password_2",
	}
	var err error

	user.GUID, err = db.CreateUser(ctx, user)
	assert.NoError(t, err)

	err = db.ChangePassword(ctx, user.GUID, "wrong_password", "new_password")
	assert.Error(t, err)
}

func UpdateUser_InvalidGUID(t *testing.T, db *database.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	err := db.UpdateUser(ctx, models.User{})
	assert.Error(t, err)
}

func UpdateUser_Success(t *testing.T, db *database.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	user := models.User{
		Login:    "test_user_3",
		Password: "test_password_3",
	}
	var err error

	user.GUID, err = db.CreateUser(ctx, user)
	assert.NoError(t, err)

	user.Age = 25
	user.Weight = 70
	user.Height = 175

	err = db.UpdateUser(ctx, user)
	assert.NoError(t, err)

	updatedUser, err := db.GetUserByGUID(ctx, user.GUID)

	assert.NoError(t, err)
	assert.Equal(t, user.GUID, updatedUser.GUID)
	assert.Equal(t, user.Login, updatedUser.Login)
	assert.Equal(t, user.Age, updatedUser.Age)
	assert.Equal(t, user.Weight, updatedUser.Weight)
	assert.Equal(t, user.Height, updatedUser.Height)
}

func setupTestDB() (*database.Database, func(), error) {
	Name := "testdb"
	User := "test"
	Password := "pass"

	ctx := context.Background()

	pgContainer, err := postgres.Run(
		ctx,
		"postgres:15-alpine",
		postgres.WithDatabase(Name),
		postgres.WithUsername(User),
		postgres.WithPassword(Password),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, nil, err
	}

	host, err := pgContainer.Host(ctx)
	if err != nil {
		return nil, nil, err
	}

	port, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		return nil, nil, err
	}

	p, err := psql.NewPSQL(ctx, psql.PSQLConfig{
		Host:            host,
		Port:            port.Int(),
		User:            User,
		Password:        Password,
		DBName:          Name,
		SSLMode:         "disable",
		MaxConns:        10,
		MinConns:        0,
		MaxConnIdleTime: 30,
	})

	db := database.NewDatabase(p)

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, nil, fmt.Errorf("cannot get current file path")
	}

	dir := filepath.Dir(filename)
	migrationsPath := filepath.Join(dir, "migrations")

	err = db.RunMigrations(ctx, migrationsPath)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		db.Close()
		if err := pgContainer.Terminate(ctx); err != nil {
			log.Panic(err)
		}
	}

	return db, cleanup, nil
}
