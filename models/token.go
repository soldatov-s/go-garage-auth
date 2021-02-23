package models

import (
	"github.com/soldatov-s/go-garage/types"
)

type Token struct {
	Signature string         `db:"signature"`
	Subject   string         `db:"subject"`
	Meta      types.NullMeta `db:"meta"`
	ExpiredAt types.NullTime `db:"expired_at"`
}

func (s *Token) SQLParamsRequest() []string {
	return []string{
		"signature",
		"subject",
		"meta",
		"expired_at",
	}
}

type TokenIntrospection struct {
	Active    bool                   `json:"active"`
	Subject   string                 `json:"subject,omitempty"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
	ExpiredAt int64                  `json:"expired_at,omitempty"`
}

type TokenAndUser struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}
