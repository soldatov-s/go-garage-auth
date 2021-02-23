package userv1

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"time"

	authv1 "github.com/soldatov-s/go-garage-auth/domains/auth/v1"
	"github.com/soldatov-s/go-garage-auth/models"
	"github.com/soldatov-s/go-garage/providers/httpsrv"
	"github.com/soldatov-s/go-garage/providers/httpsrv/echo"
	echoSwagger "github.com/soldatov-s/go-swagger/echo-swagger"
)

func (u *UserV1) userPostHandler(ec echo.Context) (err error) {
	// Swagger
	if echoSwagger.IsBuildingSwagger(ec) {
		err = errors.New("error")
		echoSwagger.AddToSwagger(ec).
			SetProduces("application/json").
			SetDescription("Create User Handler").
			SetSummary("This handler create new user").
			AddInBodyParameter("user_creds", "User creds", models.NewCredentials{}, true).
			AddResponse(http.StatusOK, "User Data", &UserDataResult{Body: models.User{}}).
			AddResponse(http.StatusBadRequest, "BAD REQUEST", httpsrv.BadRequest(err)).
			AddResponse(http.StatusConflict, "CREATE USER FAILED", httpsrv.CreateFailed(err)).
			AddResponse(http.StatusNotAcceptable, "EMAIL IS OCCUPIED", EmailIsOccupied())

		return nil
	}

	// Main code of handler
	log := echo.GetLog(ec)

	var userCreds models.NewCredentials

	var bodyBytes []byte
	if ec.Request().Body != nil {
		bodyBytes, err = ioutil.ReadAll(ec.Request().Body)

		ec.Request().Body.Close()

		if err != nil {
			log.Err(err).Msg("BAD REQUEST")

			return echo.BadRequest(ec, err)
		}
	}

	err = json.Unmarshal(bodyBytes, &userCreds)
	if err != nil {
		log.Err(err).Msg("BAD REQUEST")

		return echo.BadRequest(ec, err)
	}

	if !userCreds.Validate() {
		log.Err(err).Msg("BAD REQUEST")

		return echo.BadRequest(ec, err)
	}

	userData, err := u.createUser(&userCreds)
	if err != nil {
		if err == ErrLoginOrEmailIsOccupied {
			log.Err(err).Msgf("EMAIL IS OCCUPIED %s", &userCreds)

			return ec.JSON(
				http.StatusNotAcceptable,
				EmailIsOccupied(),
			)
		}

		log.Err(err).Msgf("CREATE USER FAILED %s", &userCreds)
		return echo.CreateFailed(ec, err)
	}

	return echo.OK(ec, UserDataResult{Body: userData})
}

func (u *UserV1) userGetHandler(ec echo.Context) (err error) {
	// Swagger
	if echoSwagger.IsBuildingSwagger(ec) {
		err = fmt.Errorf("error")
		echoSwagger.AddToSwagger(ec).
			SetProduces("application/json").
			SetDescription("Get User Handler").
			SetSummary("This handler get user data by user_id").
			AddInPathParameter("id", "User id", reflect.Int64).
			AddResponse(http.StatusOK, "User Data", &UserDataResult{Body: models.User{}}).
			AddResponse(http.StatusBadRequest, "BAD REQUEST", httpsrv.BadRequest(err)).
			AddResponse(http.StatusNotFound, "NOT FOUND DATA", httpsrv.NotFound(err))

		return nil
	}

	// Main code of handler
	log := echo.GetLog(ec)

	userID, err := strconv.ParseInt(ec.Param("id"), 10, 64)
	if err != nil {
		log.Err(err).Msgf("BAD REQUEST, id %s", ec.Param("id"))

		return echo.BadRequest(ec, err)
	}

	userData, err := u.GetUserDataByID(userID)
	if err != nil {
		log.Err(err).Msgf("NOT FOUND, id %d", userID)

		return echo.NotFound(ec, err)
	}

	return echo.OK(ec, UserDataResult{Body: userData})
}

func (u *UserV1) userPutHandler(ec echo.Context) (err error) {
	// Swagger
	if echoSwagger.IsBuildingSwagger(ec) {
		err = fmt.Errorf("error")
		echoSwagger.AddToSwagger(ec).
			SetProduces("application/json").
			SetDescription("Update User Handler").
			SetSummary("This handler update user data by user_id").
			AddInBodyParameter("user_data", "User data", &models.User{}, true).
			AddInPathParameter("id", "User id", reflect.Int64).
			AddResponse(http.StatusOK, "User Data", &UserDataResult{Body: models.User{}}).
			AddResponse(http.StatusBadRequest, "BAD REQUEST", httpsrv.BadRequest(err)).
			AddResponse(http.StatusConflict, "DATA NOT UPDATED", httpsrv.NotUpdated(err)).
			AddResponse(http.StatusNotFound, "NOT FOUND DATA", httpsrv.NotFound(err)).
			AddResponse(http.StatusNotAcceptable, "EMAIL IS OCCUPIED", EmailIsOccupied())

		return nil
	}

	// Main code of handler
	log := echo.GetLog(ec)

	userID, err := strconv.ParseInt(ec.Param("id"), 10, 64)
	if err != nil {
		log.Err(err).Msgf("BAD REQUEST, id %s", ec.Param("id"))

		return echo.BadRequest(ec, err)
	}

	var bodyBytes []byte
	if ec.Request().Body != nil {
		bodyBytes, err = ioutil.ReadAll(ec.Request().Body)

		ec.Request().Body.Close()

		if err != nil {
			log.Err(err).Msgf("USER DATA NOT UPDATED, id %d", userID)
			return echo.BadRequest(ec, err)
		}
	}

	userData, err := u.updateUserByID(userID, &bodyBytes)
	if err != nil {
		if err == ErrLoginOrEmailIsOccupied {
			log.Err(err).Msgf("EMAIL IS OCCUPIED, id %d, body %s", userID, string(bodyBytes))

			return ec.JSON(
				http.StatusNotAcceptable,
				EmailIsOccupied(),
			)
		}

		log.Err(err).Msgf("BAD REQUEST, id %d, body %s", userID, string(bodyBytes))

		return echo.NotUpdated(ec, err)
	}

	return echo.OK(ec, UserDataResult{Body: userData})
}

func (u *UserV1) credsPutHandler(ec echo.Context) (err error) {
	// Swagger
	if echoSwagger.IsBuildingSwagger(ec) {
		err = fmt.Errorf("error")
		echoSwagger.AddToSwagger(ec).
			SetProduces("application/json").
			SetDescription("Update User Credentials Handler").
			SetSummary("This handler update user credentials data by user_id").
			AddInBodyParameter("user_creds", "User creds", &models.UpdateCredentials{}, true).
			AddInPathParameter("id", "User id", reflect.Int64).
			AddResponse(http.StatusOK, "User Data", &UserDataResult{Body: models.User{}}).
			AddResponse(http.StatusBadRequest, "BAD REQUEST", httpsrv.BadRequest(err)).
			AddResponse(http.StatusConflict, "DATA NOT UPDATED", httpsrv.NotUpdated(err)).
			AddResponse(http.StatusConflict, "NEW PASSWORD SAME AS OLD", NewPasswordIsSameAsOld())

		return nil
	}

	// Main code of handler
	log := echo.GetLog(ec)

	userID, err := strconv.ParseInt(ec.Param("id"), 10, 64)
	if err != nil {
		log.Err(err).Msgf("BAD REQUEST, id %s", ec.Param("id"))

		return echo.BadRequest(ec, err)
	}

	var userCreds models.UpdateCredentials

	err = ec.Bind(&userCreds)
	if err != nil {
		log.Err(err).Msgf("BAD REQUEST, id %d", userID)

		return echo.BadRequest(ec, err)
	}

	if !userCreds.Validate() {
		log.Err(err).Msgf("BAD REQUEST, id %d, userCreds %s", userID, &userCreds)

		return echo.BadRequest(ec, err)
	}

	userData, err := u.updateUserCredsByID(userID, &userCreds)
	if err != nil {
		if err == ErrNewPasswordIsSameAsOld {
			log.Err(err).Msgf("NEW PASSWORD SAME AS OLD, id %d, userCreds %s", userID, &userCreds)

			return ec.JSON(
				http.StatusConflict,
				NewPasswordIsSameAsOld(),
			)
		}

		log.Err(err).Msgf("DATA NOT UPDATED, id %d, userCreds %s", userID, &userCreds)

		return echo.NotUpdated(ec, err)
	}

	return echo.OK(ec, UserDataResult{Body: userData})
}

func (u *UserV1) credsPostHandler(ec echo.Context) (err error) {
	// Swagger
	if echoSwagger.IsBuildingSwagger(ec) {
		err = fmt.Errorf("error")
		echoSwagger.AddToSwagger(ec).
			SetProduces("application/json").
			SetDescription("Check User Handler").
			SetSummary("This handler check user credentials. If there is a login, a login is taken; if there is no login, an email is taken").
			AddInBodyParameter("user_creds", "User creds", &models.Credentials{}, true).
			AddResponse(http.StatusOK, "User Data", &UserDataResult{Body: models.User{}}).
			AddResponse(http.StatusBadRequest, "BAD REQUEST", httpsrv.BadRequest(err)).
			AddResponse(http.StatusUnauthorized, "UNAUTHORIZED", httpsrv.Unauthorized(err))

		return nil
	}

	// Main code of handler
	log := echo.GetLog(ec)

	var userCreds models.Credentials

	err = ec.Bind(&userCreds)
	if err != nil {
		log.Err(err).Msg("BAD REQUEST")

		return echo.BadRequest(ec, err)
	}

	if !userCreds.Validate() {
		log.Err(err).Msgf("BAD REQUEST, userCreds %s", &userCreds)

		return echo.BadRequest(ec, err)
	}

	userData, err := u.GetUserDataByCreds(&userCreds)
	if err != nil {
		log.Err(err).Msgf("UNAUTHORIZED, userCreds %s", &userCreds)
		return echo.Unauthorized(ec, err)
	}

	authV1, err := authv1.Get(u.ctx)
	if err != nil {
		log.Err(err).Msg("failed to get authv1 domain")
		return echo.InternalServerError(ec, err)
	}

	if !userCreds.Validate() {
		log.Err(err).Msgf("failed to create token, userCreds %s", &userCreds)
		return echo.InternalServerError(ec, err)
	}

	token, err := authV1.CreateToken(int(userData.ID))
	if err != nil {
		log.Err(err).Msgf("CREATE SESSION FAILED %+v", &userCreds)
		return echo.BadRequest(ec, err)
	}

	cookie := new(http.Cookie)
	cookie.Name = authv1.SessionCookie
	cookie.Value = token
	cookie.Expires = time.Now().Add(u.cfg.Token.TTL)

	return echo.OK(ec, TokenAndUserResult{Body: models.TokenAndUser{Token: token, User: userData}})
}

func (u *UserV1) userDeleteHandler(ec echo.Context) (err error) {
	// Swagger
	if echoSwagger.IsBuildingSwagger(ec) {
		err = fmt.Errorf("error")
		echoSwagger.AddToSwagger(ec).
			SetProduces("application/json").
			SetDescription("Soft/Hard Delete User Handler").
			SetSummary("This handler for soft/hard delete user data by user_id").
			AddInPathParameter("id", "User id", reflect.Int64).
			AddInQueryParameter("hard", "Hard delete user, if equal true, delete hard", reflect.Bool, false).
			AddResponse(http.StatusOK, "OK", httpsrv.OkResult()).
			AddResponse(http.StatusBadRequest, "BAD REQUEST", httpsrv.BadRequest(err)).
			AddResponse(http.StatusConflict, "DATA NOT DELETED", httpsrv.NotDeleted(err))

		return nil
	}

	// Main code of handler
	log := echo.GetLog(ec)

	userID, err := strconv.ParseInt(ec.Param("id"), 10, 64)
	if err != nil {
		log.Err(err).Msgf("BAD REQUEST, id %s", ec.Param("id"))

		return echo.BadRequest(ec, err)
	}

	hard := ec.QueryParam("hard")
	if hard == "true" {
		err = u.hardDeleteUserByID(userID)
	} else {
		err = u.softDeleteUserByID(userID)
	}

	if err != nil {
		log.Err(err).Msgf("DATA NOT DELETED, id %d", userID)

		return echo.NotDeleted(ec, err)
	}

	return ec.JSON(
		http.StatusOK,
		httpsrv.OkResult(),
	)
}

func (u *UserV1) userSearchPostHandler(ec echo.Context) (err error) {
	// Swagger
	if echoSwagger.IsBuildingSwagger(ec) {
		err = fmt.Errorf("error")
		echoSwagger.AddToSwagger(ec).
			SetProduces("application/json").
			SetDescription("Find User by email Handler").
			SetSummary("This handler find user data by any field in User data struct. Can be multiple structs in request. Search by user_meta not work!").
			AddInBodyParameter("users_data", "Users data", &ArrayOfUserData{}, true).
			AddResponse(http.StatusOK, "Users data", &UsersDataResult{Body: ArrayOfUserData{}}).
			AddResponse(http.StatusBadRequest, "BAD REQUEST", httpsrv.BadRequest(err)).
			AddResponse(http.StatusNotFound, "NOT FOUND DATA", httpsrv.NotFound(err))

		return nil
	}

	// Main code of handler
	log := echo.GetLog(ec)
	var req ArrayOfMapInterface

	err = ec.Bind(&req)
	if err != nil {
		log.Err(err).Msg("BAD REQUEST")

		return echo.BadRequest(ec, err)
	}

	foundUsersData, err := u.getUserDataByUserData(&req)
	if err != nil {
		log.Err(err).Msgf("NOT FOUND DATA, request %+v", req)

		return echo.NotFound(ec, err)
	}

	return ec.JSON(
		http.StatusOK,
		UserDataResult{Body: foundUsersData},
	)
}
