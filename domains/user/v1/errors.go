package userv1

import (
	"errors"
	"net/http"

	"github.com/soldatov-s/go-garage/providers/httpsrv"
)

var (
	ErrLoginOrEmailIsOccupied = errors.New("login or email is occupied")
	ErrNewPasswordIsSameAsOld = errors.New("new password is same as old")
	ErrKeyDoNotMatch          = errors.New("key do not match")
	ErrFailedTypeCast         = errors.New("failed typecast")
)

func EmailIsOccupied() httpsrv.ErrorAnsw {
	return httpsrv.NewErrorAnsw(http.StatusNotAcceptable, "email is occupied", errors.New("email is occupied"))
}

func NewPasswordIsSameAsOld() httpsrv.ErrorAnsw {
	return httpsrv.NewErrorAnsw(http.StatusConflict, "new password is same as old", ErrNewPasswordIsSameAsOld)
}
