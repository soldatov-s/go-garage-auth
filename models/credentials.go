package models

import (
	goGarageAuthTypes "github.com/soldatov-s/go-garage-auth/types"
	"github.com/soldatov-s/go-garage/utils"
)

// Credentials is a struct for check user
type Credentials struct {
	Password string `json:"user_password"`
	Login    string `json:"user_login"`
	Email    string `json:"user_email"`
	Phone    string `json:"user_phone"`
}

func (c *Credentials) String() string {
	return utils.JoinStrings(",", []string{
		"Email: " + c.Email,
		"Phone: " + c.Phone,
		"Login: " + c.Login,
	}...)
}

func (c *Credentials) Validate() bool {
	if c.Password == "" || (c.Phone == "" && c.Email == "" && c.Login == "") {
		return false
	}

	return true
}

// NewCredentials is a stuct for create user
type NewCredentials struct {
	Credentials
	Role   goGarageAuthTypes.Role   `json:"user_role" swagtype:"string"`
	Status goGarageAuthTypes.Status `json:"user_status" swagtype:"string"`
}

func (c *NewCredentials) String() string {
	return utils.JoinStrings(",", []string{
		"Email: " + c.Email,
		"Phone: " + c.Phone,
		"Login: " + c.Login,
		"Role: " + c.Role.String(),
		"Status: " + c.Status.String(),
	}...)
}

// UpdateCredentials is a structs for update user's Credentials
type UpdateCredentials struct {
	Password    string `json:"user_password"`
	OldPassword string `json:"user_old_password"`
}

func (c *UpdateCredentials) Validate() bool {
	return c.Password != "" && c.OldPassword != ""
}
