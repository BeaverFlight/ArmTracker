package models

import (
	"pkg/roles"
	"time"

	"github.com/google/uuid"
)

type RequestAuthorization struct {
	GUID uuid.UUID  `json:"guid"`
	Role roles.Role `json:"role"`
}

type RefreshData struct {
	Refresh    uuid.UUID
	Guid       uuid.UUID
	PairID     uuid.UUID
	Role       roles.Role
	RefreshTTL time.Duration
}
