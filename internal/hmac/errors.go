package hmac

import "errors"

var (
	ErrTokenSignatureMismatch = errors.New("token signature mismatch")
	ErrInvalidTokenFormat     = errors.New("invalid token")
)
