package database

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"pkg/psql"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrLoginBusy       = errors.New("данный логин уже занят")
	ErrLoginNotFound   = errors.New("неправильный логин или пароль")
	ErrInvalidPassword = errors.New("неправильный логин или пароль")
	ErrGUIDNotFound    = errors.New("пользователь не найден")
	ErrNoUpdateFields  = errors.New("нет полей для обновления")
)

type Database struct {
	pool *psql.PSQL
}

type User struct {
	GUID             uuid.UUID `db:"id"`
	Login            string    `db:"login"`
	Password         string    `db:"password"`
	Name             string    `db:"name"`
	Height           int       `db:"height"`
	Weight           int       `db:"weight"`
	Age              int       `db:"age"`
	Admin            bool      `db:"admin"`
	RegistrationDate time.Time `db:"registration_date"`
}

func NewDatabase(psql *psql.PSQL) (*Database, error) {
	return &Database{psql}, nil
}

func (d *Database) Close() {
	d.pool.Close()
}

func (d *Database) CreateUser(ctx context.Context, user User) (uuid.UUID, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, fmt.Errorf("dbwork/CreateUser generateHashPassword: %w", err)
	}

	createQuery := "INSERT INTO users (login, password, name, height, weight, age, admin, registration_date) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"
	user.Password = string(hashPassword)
	date := time.Now()
	guid := uuid.Nil
	err = d.pool.Pool.QueryRow(ctx, createQuery, user.Login, user.Password, user.Name, user.Height, user.Weight, user.Age, user.Admin, date).Scan(&guid)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			return uuid.Nil, ErrLoginBusy
		}
		return uuid.Nil, fmt.Errorf("dbwork/CreateUser insert: %w", err)
	}

	return guid, nil
}

func (d *Database) GetUserByLogin(ctx context.Context, login string) (User, error) {
	selectQuery := "SELECT id, login, password, name, height, weight, age, admin, registration_date FROM users WHERE login = $1"
	var user User

	err := d.pool.Pool.QueryRow(ctx, selectQuery, login).Scan(&user.GUID, &user.Login, &user.Password, &user.Name, &user.Height, &user.Weight, &user.Age, &user.Admin, &user.RegistrationDate)

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return User{}, fmt.Errorf("dbwork/GetUserByLogin: %w", ErrLoginNotFound)
		}
		return User{}, fmt.Errorf("dbwork/GetUserByLogin: %w", err)
	}

	user.Password = ""
	return user, nil
}

func (d *Database) GetUserByGUID(ctx context.Context, guid uuid.UUID) (User, error) {
	selectQuery := "SELECT id, login, password, name, height, weight, age, admin, registration_date FROM users WHERE id = $1"
	var user User

	err := d.pool.Pool.QueryRow(ctx, selectQuery, guid.String()).Scan(&user.GUID, &user.Login, &user.Password, &user.Name, &user.Height, &user.Weight, &user.Age, &user.Admin, &user.RegistrationDate)

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return User{}, fmt.Errorf("dbwork/GetUserByGUID: %w", ErrGUIDNotFound)
		}
		return User{}, fmt.Errorf("dbwork/GetUserByGUID: %v", err)
	}

	user.Password = ""
	return user, nil
}

func (d *Database) VerifyUser(ctx context.Context, login string, password string) (*User, error) {
	selectQuery := "SELECT id, login, password,  admin FROM users WHERE login = $1"
	var user User

	err := d.pool.Pool.QueryRow(ctx, selectQuery, login).Scan(&user.GUID, &user.Login, &user.Password, &user.Admin)

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, fmt.Errorf("dbwork/VerifyUser: %w", ErrLoginNotFound)
		}
		return nil, fmt.Errorf("dbwork/VerifyUser: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidPassword
	}

	user.Password = ""
	return &user, nil
}

func (d *Database) MakeAdmin(ctx context.Context, guid uuid.UUID) (uuid.UUID, error) {
	updateQuery := `UPDATE users
	                       SET admin = true
						   WHERE id=$1`

	_, err := d.pool.Pool.Exec(ctx, updateQuery, guid.String())
	if err != nil {
		return uuid.Nil, err
	}

	return guid, nil
}

func (d *Database) UpdateUser(ctx context.Context, user User) error {
	updateQuery := `UPDATE users
	                       SET `

	query, args, index := d.selectParameters(user, ", ")
	updateQuery += query

	if index == 1 {
		return ErrNoUpdateFields
	}

	updateQuery += fmt.Sprintf(` WHERE id=$%d`, index)
	args = append(args, user.GUID)

	_, err := d.pool.Pool.Exec(ctx, updateQuery, args...)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) selectParameters(user User, sep string) (string, []any, int) {
	query := ""
	args := []any{}
	index := 1
	if !user.RegistrationDate.IsZero() {
		query += fmt.Sprintf(`registration_date=$%d`, index)
		args = append(args, user.RegistrationDate)
		index++
	}

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
		return fmt.Errorf("dbwork/ChangePassword: %w", ErrInvalidPassword)
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("dbwork/ChangePassword: %w", err)
	}

	_, err = d.pool.Pool.Exec(ctx, "UPDATE users SET password=$1 WHERE id=$2", hashPassword, userID)
	if err != nil {
		return fmt.Errorf("dbwork/ChangePassword: %w", err)
	}

	return nil
}

func (d *Database) getUserByGUIDWithPassword(ctx context.Context, guid uuid.UUID) (User, error) {
	selectQuery := "SELECT id, login, password, name, height, weight, age, admin, registration_date FROM users WHERE id = $1"
	var user User

	err := d.pool.Pool.QueryRow(ctx, selectQuery, guid.String()).Scan(&user.GUID, &user.Login, &user.Password, &user.Name, &user.Height, &user.Weight, &user.Age, &user.Admin, &user.RegistrationDate)

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return User{}, fmt.Errorf("dbwork/GetUserByGUID: %w", ErrGUIDNotFound)
		}
		return User{}, fmt.Errorf("dbwork/GetUserByGUID: %w", err)
	}

	return user, nil
}
