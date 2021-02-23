package authv1

import (
	"strconv"
	"time"

	"github.com/soldatov-s/go-garage-auth/internal/hmac"
	"github.com/soldatov-s/go-garage-auth/models"
	"github.com/soldatov-s/go-garage/providers/db"
	"github.com/soldatov-s/go-garage/x/sql"
)

func (a *AuthV1) CreateToken(id int) (string, error) {
	var request models.Token

	strategy, err := hmac.Get(a.ctx)
	if err != nil {
		return "", err
	}

	token, sign, err := strategy.Generate()
	if err != nil {
		return "", err
	}

	request.Signature = sign
	request.Subject = strconv.Itoa(id)
	request.ExpiredAt.SetTime(time.Now().Add(a.cfg.Token.TTL))

	if a.db.Conn == nil {
		return "", db.ErrDBConnNotEstablished
	}

	_, err = sql.InsertInto(a.db.Conn, "production.token", &request)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (a *AuthV1) GetToken(id string) (data *models.Token, err error) {
	data = &models.Token{}

	if a.db.Conn == nil {
		return nil, db.ErrDBConnNotEstablished
	}

	err = a.db.Conn.Get(data, "select * from production.token where signature=$1", id)
	if err != nil {
		return nil, err
	}

	a.log.Debug().Msgf("session %+v", data)

	return
}

func (a *AuthV1) DeleteToken(id string) (err error) {
	if a.db.Conn == nil {
		return db.ErrDBConnNotEstablished
	}

	_, err = a.db.Conn.Exec(a.db.Conn.Rebind("DELETE FROM production.token WHERE signature=$1"), id)

	if err != nil {
		return err
	}

	return nil
}
