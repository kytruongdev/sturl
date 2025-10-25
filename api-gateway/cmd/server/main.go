package main

import (
	"context"
	"log"

	"github.com/kytruongdev/sturl/api-gateway/internal/handler"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/app"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/httpserver"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/logger"
)

func main() {
	cfg, err := initAppConfig()

	rootLog := logger.New(cfg.ServerCfg.ServiceName, cfg.ServerCfg.LogLevel, cfg.ServerCfg.AppEnv)
	rootCtx := logger.ToContext(context.Background(), rootLog)
	l := logger.FromContext(rootCtx)

	l.Info().Msg("Starting app initialization")

	if err != nil {
		log.Fatal(err)
	}

	rtr, err := initRouter()
	if err != nil {
		log.Fatal(err)
	}

	l.Info().Msg("App initialization completed")

	httpserver.Start(httpserver.Handler(rootCtx, httpserver.NewCORSConfig(rtr.CorsOrigins), rtr.Routes),
		cfg.ServerCfg)
}

func initAppConfig() (app.Config, error) {
	cfg := app.NewConfig()

	return cfg, cfg.Validate()
}

func initRouter() (handler.Router, error) {
	return handler.Router{
		CorsOrigins: []string{"*"},
	}, nil
}
