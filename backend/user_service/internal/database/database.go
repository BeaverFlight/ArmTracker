package database

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"user_service/internal/models"

	"pkg/psql"
	"pkg/roles"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type Database struct {
	pool *psql.PSQL
}

func NewDatabase(psql *psql.PSQL) *Database {
	return &Database{psql}
}

func (d *Database) Close() {
	d.pool.Close()
}

func (d *Database) RunMigrations(ctx context.Context, migrationsPath string) error {
	return d.pool.RunMigrations(ctx, migrationsPath)
}

func (d *Database) CreateUser(ctx context.Context, user models.User) (uuid.UUID, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, fmt.Errorf("Database/CreateUser generateHashPassword: %w", err)
	}

	createQuery := "INSERT INTO users (login, password, name, height, weight, age, role, registration_date, email) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id"
	user.Password = string(hashPassword)
	date := time.Now()
	guid := uuid.Nil
	err = d.pool.Pool.QueryRow(ctx, createQuery, user.Login, user.Password, user.Name, user.Height, user.Weight, user.Age, user.Role, date, user.Email).Scan(&guid)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			return uuid.Nil, models.ErrLoginBusy
		}
		return uuid.Nil, fmt.Errorf("Database/CreateUser insert: %w", err)
	}

	return guid, nil
}

func (d *Database) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	selectQuery := "SELECT id, login, password, name, height, weight, age, role, registration_date, email FROM users WHERE login = $1"
	var user models.User

	err := d.pool.Pool.QueryRow(ctx, selectQuery, login).Scan(&user.GUID, &user.Login, &user.Password, &user.Name, &user.Height, &user.Weight, &user.Age, &user.Role, &user.RegistrationDate, &user.Email)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, fmt.Errorf("Database/GetUserByLogin: %w", models.ErrLoginNotFound)
	}

	if err != nil {
		return models.User{}, fmt.Errorf("Database/GetUserByLogin: %w", err)
	}

	user.Password = ""
	return user, nil
}

func (d *Database) GetUserByGUID(ctx context.Context, guid uuid.UUID) (models.User, error) {
	selectQuery := "SELECT id, login, password, name, height, weight, age, role, registration_date, email FROM users WHERE id = $1"
	var user models.User

	err := d.pool.Pool.QueryRow(ctx, selectQuery, guid.String()).Scan(&user.GUID, &user.Login, &user.Password, &user.Name, &user.Height, &user.Weight, &user.Age, &user.Role, &user.RegistrationDate, &user.Email)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, fmt.Errorf("Database/GetUserByGUID: %w", models.ErrGUIDNotFound)
	}
	if err != nil {
		return models.User{}, fmt.Errorf("Database/GetUserByGUID: %w", err)
	}

	user.Password = ""
	return user, nil
}

func (d *Database) VerifyUser(ctx context.Context, login, password string) (uuid.UUID, roles.Role, error) {
	selectQuery := "SELECT id, login, password,  role FROM users WHERE login = $1"
	var user models.User

	err := d.pool.Pool.QueryRow(ctx, selectQuery, login).Scan(&user.GUID, &user.Login, &user.Password, &user.Role)

	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, roles.RoleNone, fmt.Errorf("Database/VerifyUser: %w", models.ErrLoginNotFound)
	}
	if err != nil {
		return uuid.Nil, roles.RoleNone, fmt.Errorf("Database/VerifyUser: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return uuid.Nil, roles.RoleNone, fmt.Errorf("Database/VerifyUser: %w", models.ErrInvalidPassword)
	}
	user.Password = ""
	return user.GUID, user.Role, nil
}

func (d *Database) SetRole(ctx context.Context, guid uuid.UUID, role roles.Role) error {
	updateQuery := `UPDATE users
	                       SET role = $1
						   WHERE id=$2`

	result, err := d.pool.Pool.Exec(ctx, updateQuery, role, guid.String())
	if err != nil {
		return fmt.Errorf("Database/MakeAdmin: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("Database/MakeAdmin: %w", models.ErrGUIDNotFound)
	}

	return nil
}

func (d *Database) UpdateUser(ctx context.Context, user models.User) error {
	updateQuery := `UPDATE users
	                       SET `

	query, args, index := d.selectParameters(user, ", ")
	updateQuery += query

	if index == 1 {
		return fmt.Errorf("Database/UpdateUser: %w", models.ErrNoUpdateFields)
	}

	updateQuery += fmt.Sprintf(` WHERE id=$%d`, index)
	args = append(args, user.GUID)

	_, err := d.pool.Pool.Exec(ctx, updateQuery, args...)
	if err != nil {
		return fmt.Errorf("Database/UpdateUser: %w", err)
	}

	return nil
}

func (d *Database) selectParameters(user models.User, sep string) (string, []any, int) {
	query := ""
	args := []any{}
	index := 1

	if user.Name != "" {
		if index != 1 {
			query += sep
		}
		query += fmt.Sprintf("name=$%d", index)
		args = append(args, user.Name)
		index++
	}

	if user.Height != 0 {
		if index != 1 {
			query += sep
		}
		query += fmt.Sprintf("height=$%d", index)
		args = append(args, user.Height)
		index++
	}

	if user.Weight != 0 {
		if index != 1 {
			query += sep
		}
		query += fmt.Sprintf("weight=$%d", index)
		args = append(args, user.Weight)
		index++
	}

	if user.Age != 0 {
		if index != 1 {
			query += sep
		}
		query += fmt.Sprintf("age=$%d", index)
		args = append(args, user.Age)
		index++
	}

	return query, args, index
}

func (d *Database) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := d.getUserByGUIDWithPassword(ctx, userID)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))

	if err != nil {
		return fmt.Errorf("Database/ChangePassword: %w", models.ErrInvalidPassword)
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("Database/ChangePassword: %w", err)
	}

	_, err = d.pool.Pool.Exec(ctx, "UPDATE users SET password=$1 WHERE id=$2", hashPassword, userID)
	if err != nil {
		return fmt.Errorf("Database/ChangePassword: %w", err)
	}

	return nil
}

func (d *Database) getUserByGUIDWithPassword(ctx context.Context, guid uuid.UUID) (models.User, error) {
	selectQuery := "SELECT id, login, password, name, height, weight, age, role, registration_date, email FROM users WHERE id = $1"
	var user models.User

	err := d.pool.Pool.QueryRow(ctx, selectQuery, guid.String()).Scan(&user.GUID, &user.Login, &user.Password, &user.Name, &user.Height, &user.Weight, &user.Age, &user.Role, &user.RegistrationDate, &user.Email)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, fmt.Errorf("Database/GetUserByGUID: %w", models.ErrGUIDNotFound)
	}
	if err != nil {
		return models.User{}, fmt.Errorf("Database/GetUserByGUID: %w", err)
	}

	return user, nil
}
