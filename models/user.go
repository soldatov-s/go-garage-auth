package models

import (
	"encoding/json"
	"errors"

	goGarageAuthTypes "github.com/soldatov-s/go-garage-auth/types"
	"github.com/soldatov-s/go-garage/models"
	"github.com/soldatov-s/go-garage/types"
)

var (
	ErrEmptyMail = errors.New("empty email")
)

type User struct {
	ID             int64                    `json:"user_id" db:"user_id"`
	Hash           string                   `json:"-" db:"user_hash"`
	Login          string                   `json:"user_login" db:"user_login"`
	Email          string                   `json:"user_email" db:"user_email"`
	Phone          string                   `json:"user_phone" db:"user_phone"`
	Status         goGarageAuthTypes.Status `json:"user_status" db:"user_status" swagtype:"string"`
	Role           goGarageAuthTypes.Role   `json:"user_role" db:"user_role" swagtype:"string"`
	Meta           types.NullMeta           `json:"user_meta" db:"user_meta"`
	ActivationHash types.NullString         `json:"-" db:"user_activation_hash"`
	models.Timestamp
}

func (u *User) SQLParamsRequest() []string {
	return []string{
		"user_hash",
		"user_login",
		"user_email",
		"user_phone",
		"user_status",
		"user_role",
		"user_meta",
		"user_activation_hash",
		"created_at",
		"updated_at",
		"deleted_at",
	}
}

func (u *User) MarshalBinary() ([]byte, error) {
	return json.Marshal(u)
}

func (u *User) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, u)
}

func (u *User) Validate() error {
	if u.Email == "" {
		return ErrEmptyMail
	}

	return nil
}
