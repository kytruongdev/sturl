package main

import (
	"context"
	"log"

	"github.com/kytruongdev/sturl/api-gateway/internal/handler"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/app"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/env"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/httpserver"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/logger"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/proxy"
)

func main() {
	cfg := initAppConfig()

	rootLog := logger.New(cfg.ServerCfg.ServiceName, cfg.ServerCfg.LogLevel, cfg.ServerCfg.AppEnv)
	rootCtx := logger.ToContext(context.Background(), rootLog)
	l := logger.FromContext(rootCtx)

	l.Info().Msg("Starting app initialization")

	proxyRegistry := proxy.NewRegistry()

	// Register all proxies at startup
	err := proxyRegistry.Register("url-shortener", env.GetAndValidateF("URL_SHORTENER_URL"))
	if err != nil {
		l.Error().Err(err).Msg("failed to init proxy")
	}

	rtr := initRouter()

	l.Info().Msg("App initialization completed")

	httpserver.Start(httpserver.Handler(rootCtx, httpserver.NewCORSConfig(rtr.CorsOrigins), rtr.Routes),
		cfg.ServerCfg)
}

func initAppConfig() app.Config {
	cfg := app.NewConfig()

	if err := cfg.Validate(); err != nil {
		log.Fatal("[initAppConfig] err: ", err)
	}

	return cfg
}

func initRouter() handler.Router {
	return handler.Router{
		CorsOrigins: []string{"*"},
	}
}
