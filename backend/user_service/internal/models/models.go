package models

import (
	"time"

	"github.com/google/uuid"
)

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

type UserChangePassword struct {
	OldPassword string
	NewPassword string
	GUID        uuid.UUID
}
