package cfg

import (
	"context"

	"github.com/soldatov-s/go-garage-auth/internal/hmac"
	"github.com/soldatov-s/go-garage/providers/config"
	"github.com/soldatov-s/go-garage/providers/db/pq"
	"github.com/soldatov-s/go-garage/providers/httpsrv/echo"
	"github.com/soldatov-s/go-garage/providers/logger"
	"github.com/soldatov-s/go-garage/providers/stats/garage"
)

const (
	DBName    = "auth"
	StatsName = "statsTest"

	PublicHTTP  = "public"
	PrivateHTTP = "private"
	V1          = "1"
)

type Config struct {
	Logger      *logger.Config
	DB          *pq.Config
	PublicHTTP  *echo.Config
	PrivateHTTP *echo.Config
	Stats       *garage.Config
	Token       *hmac.Config
}

func Get(ctx context.Context) *Config {
	return config.Get(ctx).Service.(*Config)
}

func NewConfig() *Config {
	return &Config{
		Logger: &logger.Config{},
		DB: &pq.Config{
			DSN: "postgres://postgres:secret@postgres:5432/auth",
			Migrate: &pq.MigrateConfig{
				Directory: "/internal/db/migrations/pg",
				Action:    "up",
			},
		},
		PublicHTTP: &echo.Config{
			Address:    "0.0.0.0:9000",
			HideBanner: true,
			HidePort:   true,
		},
		PrivateHTTP: &echo.Config{
			Address:    "0.0.0.0:9100",
			HideBanner: true,
			HidePort:   true,
		},
		Stats: &garage.Config{
			HTTPProviderName: echo.DefaultProviderName,
			HTTPEnityName:    "private",
		},
	}
}
