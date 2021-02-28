package authv1

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/soldatov-s/go-garage-auth/internal/cfg"
	"github.com/soldatov-s/go-garage/domains"
	"github.com/soldatov-s/go-garage/providers/db/pq"
	"github.com/soldatov-s/go-garage/providers/httpsrv/echo"
	"github.com/soldatov-s/go-garage/providers/logger"
)

const (
	DomainName = "authv1"
)

type empty struct{}

type AuthV1 struct {
	log   zerolog.Logger
	ctx   context.Context
	db    *pq.Enity
	cfg   *cfg.Config
	mutex *pq.Mutex
}

func Registrate(ctx context.Context) (context.Context, error) {
	a := &AuthV1{
		ctx: ctx,
		log: logger.GetPackageLogger(ctx, empty{}),
		cfg: cfg.Get(ctx),
	}
	var err error
	if a.db, err = pq.GetEnityTypeCast(ctx, cfg.DBName); err != nil {
		return nil, err
	}

	if a.mutex, err = pq.NewMutex(a.db.Conn, 0); err != nil {
		return nil, err
	}

	go a.ClearOldTokens()

	privateV1, err := echo.GetAPIVersionGroup(ctx, cfg.PrivateHTTP, cfg.V1)
	if err != nil {
		return nil, err
	}

	grProtect := privateV1.Group
	grProtect.Use(echo.HydrationLogger(&a.log))
	grProtect.POST("/auth/revoke", echo.Handler(a.revokePostHandler))
	grProtect.GET("/auth/introspect", echo.Handler(a.introspectGetHandler))

	return domains.RegistrateByName(ctx, DomainName, a), nil
}

func Get(ctx context.Context) (*AuthV1, error) {
	if v, ok := domains.GetByName(ctx, DomainName).(*AuthV1); ok {
		return v, nil
	}
	return nil, domains.ErrInvalidDomainType
}
