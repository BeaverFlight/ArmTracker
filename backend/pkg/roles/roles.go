package roles

import "github.com/jackc/pgx/v5/pgtype"

type Role string

const (
	RoleNone    Role = "None"
	RoleUser    Role = "user"
	RolePremium Role = "premium"
	RoleCoach   Role = "coach"
	RoleAdmin   Role = "admin"
)

func (r *Role) ScanText(v pgtype.Text) error {
	if !v.Valid {
		*r = RoleNone
		return nil
	}
	*r = Role(v.String)
	return nil
}

func (r Role) TextValue() (pgtype.Text, error) {
	if r == "" {
		r = RoleNone
	}
	return pgtype.Text{String: string(r), Valid: true}, nil
}

var roleWeight = map[Role]int{
	RoleNone:    0,
	RoleUser:    1,
	RolePremium: 2,
	RoleCoach:   3,
	RoleAdmin:   4,
}

func (r Role) HasAccess(required Role) bool {
	return roleWeight[r] >= roleWeight[required]
}
