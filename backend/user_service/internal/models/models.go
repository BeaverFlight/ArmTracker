package models

import (
	"pkg/roles"
	"time"

	"github.com/google/uuid"
)

type User struct {
	GUID             uuid.UUID  `db:"id" json:"guid"`
	Login            string     `db:"login" json:"login"`
	Email            string     `db:"email" json:"email" validate:"required,email"`
	Password         string     `db:"password" json:"password,omitempty"`
	Name             string     `db:"name" json:"name,omitempty"`
	Height           int        `db:"height" json:"height,omitempty"`
	Weight           int        `db:"weight" json:"weight,omitempty"`
	Age              int        `db:"age" json:"age,omitempty"`
	Role             roles.Role `db:"role" json:"role"`
	RegistrationDate time.Time  `db:"registration_date" json:"registration_date"`
}

type UserChangePassword struct {
	OldPassword string
	NewPassword string
	GUID        uuid.UUID
}
