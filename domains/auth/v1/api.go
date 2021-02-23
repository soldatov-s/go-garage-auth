package authv1

import (
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/soldatov-s/go-garage-auth/internal/hmac"
	"github.com/soldatov-s/go-garage-auth/models"
	"github.com/soldatov-s/go-garage/providers/httpsrv"
	"github.com/soldatov-s/go-garage/providers/httpsrv/echo"
	echoSwagger "github.com/soldatov-s/go-swagger/echo-swagger"
)

const (
	SessionCookie = "go-garage-session"
)

func (a *AuthV1) getTokenFromRequest(ec echo.Context) (string, error) {
	token := ec.QueryParam("token")
	if token == "" {
		sesionCookie, err := ec.Cookie(SessionCookie)
		if err != nil {
			return "", err
		}
		token = sesionCookie.Value
	}

	return token, nil
}

func (a *AuthV1) revokePostHandler(ec echo.Context) (err error) {
	// Swagger
	if echoSwagger.IsBuildingSwagger(ec) {
		err = fmt.Errorf("error")
		echoSwagger.AddToSwagger(ec).
			SetProduces("application/json").
			SetDescription("Revoke token Handler").
			SetSummary("This handler for revoking token").
			AddInQueryParameter("token", "Deleted token", reflect.Bool, false).
			AddResponse(http.StatusOK, "OK", httpsrv.OkResult()).
			AddResponse(http.StatusBadRequest, "BAD REQUEST", httpsrv.BadRequest(err)).
			AddResponse(http.StatusConflict, "DATA NOT DELETED", httpsrv.NotDeleted(err))

		return nil
	}
	// Main code of handler
	log := echo.GetLog(ec)

	token, err := a.getTokenFromRequest(ec)
	if err != nil {
		log.Err(err).Msg("getting token from request failed")
		return echo.BadRequest(ec, err)
	}

	strategy, err := hmac.Get(a.ctx)
	if err != nil {
		log.Err(err).Msgf("get token failed %s", token)
		return echo.BadRequest(ec, err)
	}

	err = a.DeleteToken(strategy.Signature(token))
	if err != nil {
		log.Err(err).Msgf("revoking token failed %s", token)

		return echo.NotDeleted(ec, err)
	}

	return echo.OkResult(ec)
}

func (a *AuthV1) introspectGetHandler(ec echo.Context) (err error) {
	// Swagger
	if echoSwagger.IsBuildingSwagger(ec) {
		err = fmt.Errorf("error")
		echoSwagger.AddToSwagger(ec).
			SetProduces("application/json").
			SetDescription("Introspect token Handler").
			SetSummary("This handler for introspection token").
			AddInQueryParameter("token", "Deleted token", reflect.Bool, false).
			AddResponse(http.StatusOK, "OK", &TokenDataResult{Body: models.Token{}}).
			AddResponse(http.StatusBadRequest, "BAD REQUEST", httpsrv.BadRequest(err)).
			AddResponse(http.StatusConflict, "DATA NOT DELETED", httpsrv.NotDeleted(err))

		return nil
	}
	// Main code of handler
	log := echo.GetLog(ec)

	token, err := a.getTokenFromRequest(ec)
	if err != nil {
		log.Err(err).Msg("getting session from request failed")
		return echo.BadRequest(ec, err)
	}

	strategy, err := hmac.Get(a.ctx)
	if err != nil {
		log.Err(err).Msgf("get token failed %s", token)
		return echo.BadRequest(ec, err)
	}

	err = strategy.Validate(token)
	if err != nil {
		log.Err(err).Msgf("token %s isn't valid", token)
		return echo.OK(ec, TokenDataResult{Body: &models.TokenIntrospection{}})
	}

	session, err := a.GetToken(strategy.Signature(token))
	if err != nil || session.ExpiredAt.Time.Before(time.Now().UTC()) {
		log.Err(err).Msgf("get token failed %s", token)
		return echo.OK(ec, TokenDataResult{Body: &models.TokenIntrospection{}})
	}

	log.Debug().Msgf("find session for subject %s", session.Subject)

	intropsectResullt := &models.TokenIntrospection{
		Active:    true,
		Subject:   session.Subject,
		Meta:      session.Meta.Map,
		ExpiredAt: session.ExpiredAt.Time.Unix(),
	}
	return echo.OK(ec, TokenDataResult{Body: intropsectResullt})
}
