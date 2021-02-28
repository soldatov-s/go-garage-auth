package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	authv1 "github.com/soldatov-s/go-garage-auth/domains/auth/v1"
	userv1 "github.com/soldatov-s/go-garage-auth/domains/user/v1"
	"github.com/soldatov-s/go-garage-auth/internal/cfg"
	"github.com/soldatov-s/go-garage-auth/internal/hmac"
	"github.com/soldatov-s/go-garage/app"
	"github.com/soldatov-s/go-garage/meta"
	"github.com/soldatov-s/go-garage/providers/config/envconfig"
	"github.com/soldatov-s/go-garage/providers/db/pq"
	"github.com/soldatov-s/go-garage/providers/httpsrv/echo"
	"github.com/soldatov-s/go-garage/providers/logger"
	"github.com/soldatov-s/go-garage/providers/stats"
	"github.com/soldatov-s/go-garage/providers/stats/garage"
	"github.com/spf13/cobra"
)

type empty struct{}

func addMetrics(ctx context.Context) error {
	s, err := garage.GetEnityTypeCast(ctx, cfg.StatsName)
	if err != nil {
		return err
	}

	if err := s.RegisterReadyCheck("TEST",
		func() (bool, string) {
			return true, "test error"
		}); err != nil {
		return err
	}

	metricOptions := &stats.MetricOptions{
		Metric: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "test_alive",
				Help: "test server link",
			}),
		MetricFunc: func(m interface{}) {
			(m.(prometheus.Gauge)).Set(1)
		},
	}

	if err := s.RegisterMetric(
		"test",
		metricOptions); err != nil {
		return err
	}

	return nil
}

func initService() context.Context {
	// Create context
	ctx := app.CreateAppContext(context.Background())

	// Set app info
	ctx = meta.SetAppInfo(ctx, appName, builded, hash, version, description)

	// Load and parse config
	ctx, err := envconfig.RegistrateAndParse(ctx, cfg.NewConfig())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("configuration parsed successfully")
	c := cfg.Get(ctx)

	// Registrate logger
	ctx = logger.RegistrateAndInitilize(ctx, c.Logger)

	// Get logger for package
	log := logger.GetPackageLogger(ctx, empty{})

	a := meta.Get(ctx)
	log.Info().Msgf("starting %s (%s)...", a.Name, a.GetBuildInfo())
	log.Info().Msg(a.Description)

	return ctx
}

func initProviders(ctx context.Context) context.Context {
	var err error
	c := cfg.Get(ctx)
	log := logger.GetPackageLogger(ctx, empty{})

	// Initialize providers
	ctx, err = pq.RegistrateEnity(ctx, cfg.DBName, c.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("failed create db connection")
	}

	// Public HTTP
	ctx, err = echo.RegistrateEnity(ctx, cfg.PublicHTTP, c.PublicHTTP)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to registrate http")
	}

	publicEchoEnity, err := echo.GetEnityTypeCast(ctx, cfg.PublicHTTP)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get http")
	}

	publicEchoEnity.Server.Use(echo.CORSDefault(), echo.HydrationRequestID())

	if err = publicEchoEnity.CreateAPIVersionGroup(cfg.V1); err != nil {
		log.Fatal().Err(err).Msg("failed to create api group")
	}

	// Private HTTP
	ctx, err = echo.RegistrateEnity(ctx, cfg.PrivateHTTP, c.PrivateHTTP)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to registrate http")
	}

	privateEchoEnity, err := echo.GetEnityTypeCast(ctx, cfg.PrivateHTTP)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get http")
	}

	privateEchoEnity.Server.Use(echo.CORSDefault(), echo.HydrationRequestID())

	if err = privateEchoEnity.CreateAPIVersionGroup(cfg.V1); err != nil {
		log.Fatal().Err(err).Msg("failed to create api group")
	}

	if ctx, err = garage.RegistrateEnity(ctx, cfg.StatsName, c.Stats); err != nil {
		log.Fatal().Err(err).Msg("failed to registrate stats")
	}

	if err = addMetrics(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to registrate metrics")
	}

	return ctx
}

func initDomains(ctx context.Context) context.Context {
	var err error
	log := logger.GetPackageLogger(ctx, empty{})

	// Initilize domains
	if ctx, err = userv1.Registrate(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to create domain authv1")
	}

	if ctx, err = authv1.Registrate(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to create domain authv1")
	}

	if ctx, err = hmac.Registrate(ctx, cfg.Get(ctx).Token.HMAC); err != nil {
		log.Fatal().Err(err).Msg("failed to create domain hmac")
	}

	return ctx
}

func serveHandler(_ *cobra.Command, _ []string) {
	ctx := initService()
	log := logger.GetPackageLogger(ctx, empty{})

	ctx = initProviders(ctx)
	ctx = initDomains(ctx)

	// Start connect
	if err := app.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to start providers")
	}

	app.Loop(ctx)
}
