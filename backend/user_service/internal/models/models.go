package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	GUID             uuid.UUID `db:"id" json:"guid"`
	Login            string    `db:"login" json:"login"`
	Password         string    `db:"password" json:"password,omitempty"`
	Name             string    `db:"name" json:"name,omitempty"`
	Height           int       `db:"height" json:"height,omitempty"`
	Weight           int       `db:"weight" json:"weight,omitempty"`
	Age              int       `db:"age" json:"age,omitempty"`
	Admin            bool      `db:"admin" json:"admin"`
	RegistrationDate time.Time `db:"registration_date" json:"registration_date"`
}

type UserChangePassword struct {
	OldPassword string
	NewPassword string
	GUID        uuid.UUID
}
