package userv1

import (
	"github.com/soldatov-s/go-garage-auth/models"
	"github.com/soldatov-s/go-garage/providers/httpsrv"
)

// Return separated items
type UserDataResult httpsrv.ResultAnsw

type TokenAndUserResult httpsrv.ResultAnsw

// Return array of items
type UsersDataResult httpsrv.ResultAnsw
type ArrayOfUserData []models.User
type ArrayOfMapInterface []map[string]interface{}
