package userv1

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/soldatov-s/go-garage-auth/internal/cfg"
	"github.com/soldatov-s/go-garage/domains"
	"github.com/soldatov-s/go-garage/providers/db/pq"
	"github.com/soldatov-s/go-garage/providers/httpsrv/echo"
	"github.com/soldatov-s/go-garage/providers/logger"
)

const (
	DomainName = "userv1"
)

type empty struct{}

type UserV1 struct {
	log zerolog.Logger
	ctx context.Context
	db  *pq.Enity
	// mutex for creating user
	mu *pq.Mutex
	// cached value of stmt for create user
	createUserStmt *sqlx.NamedStmt
	// cached value of current counter partitions
	lastID int64
	cfg    *cfg.Config
}

func Registrate(ctx context.Context) (context.Context, error) {
	u := &UserV1{
		ctx: ctx,
		log: logger.GetPackageLogger(ctx, empty{}),
		cfg: cfg.Get(ctx),
	}
	var err error
	if u.db, err = pq.GetEnityTypeCast(ctx, cfg.DBName); err != nil {
		return nil, err
	}

	u.mu, err = u.db.NewMutex(checkInterval)
	if err != nil {
		return nil, err
	}
	u.mu.GenerateLockID(cfg.DBName)

	privateV1, err := echo.GetAPIVersionGroup(ctx, cfg.PrivateHTTP, cfg.V1)
	if err != nil {
		return nil, err
	}

	grProtect := privateV1.Group
	grProtect.Use(echo.HydrationLogger(&u.log))
	grProtect.POST("/users", echo.Handler(u.userPostHandler))
	grProtect.GET("/users/:id", echo.Handler(u.userGetHandler))
	grProtect.PUT("/users/:id", echo.Handler(u.userPutHandler))
	grProtect.PUT("/credentials/:id", echo.Handler(u.credsPutHandler))
	grProtect.POST("/credentials", echo.Handler(u.credsPostHandler))
	grProtect.DELETE("/users/:id", echo.Handler(u.userDeleteHandler))
	grProtect.POST("/users/search", echo.Handler(u.userSearchPostHandler))

	return domains.RegistrateByName(ctx, DomainName, u), nil
}

func Get(ctx context.Context) (*UserV1, error) {
	if v, ok := domains.GetByName(ctx, DomainName).(*UserV1); ok {
		return v, nil
	}
	return nil, domains.ErrInvalidDomainType
}
