package models

import (
	"pkg/roles"

	"github.com/google/uuid"
)

type RequestAuthorization struct {
	GUID uuid.UUID  `json:"guid"`
	Role roles.Role `json:"role"`
}
