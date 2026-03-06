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
	LoginBusy            = errors.New("Данный логин уже занят")
	LoginNotFound        = errors.New("Неправильный логин или пароль")
	PasswordIsNotCorrect = errors.New("Неправильный логин или пароль")
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

func NewDatabase(ctx context.Context, psql *psql.PSQL) (*Database, error) {
	return &Database{psql}, nil
}

func (d *Database) Close() {
	d.pool.Close()
}

func (d *Database) CreateUser(ctx context.Context, user User) (uuid.UUID, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, fmt.Errorf("dbwork/CreateUser generateHashPassword: %v", err)
	}

	createQuery := "INSERT INTO users (login, password, name, height, weight, age, admin, registration_date) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"
	user.Password = string(hashPassword)
	date := time.Now()
	guid := uuid.Nil
	err = d.pool.Pool.QueryRow(ctx, createQuery, user.Login, user.Password, user.Name, user.Height, user.Weight, user.Age, user.Admin, date).Scan(&guid)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			return uuid.Nil, LoginBusy
		}
		return uuid.Nil, fmt.Errorf("dbwork/CreateUser insert: %v", err)
	}

	return guid, nil
}
