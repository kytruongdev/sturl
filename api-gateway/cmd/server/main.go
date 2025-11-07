package main

import (
	"context"
	"log"

	"github.com/kytruongdev/sturl/api-gateway/internal/handler"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/app"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/env"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/httpserver"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/proxy"
)

func main() {
	cfg := initAppConfig()

	// --- Setup logging monitoring
	rootCtx := context.Background()
	mon, shutdown, err := monitoring.Init(rootCtx, monitoring.ConfigFromEnv())
	if err != nil {
		log.Fatalf("init monitoring failed: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	l := mon.LoggerWithNonSpan()
	l.Info().Msg("Starting app initialization")

	// --- Setup proxies
	registerProxies(rootCtx)

	// --- Setup routers
	rtr := initRouter()

	l.Info().Msg("App initialization completed")

	// --- Start server
	httpserver.Start(httpserver.Handler(mon, httpserver.NewCORSConfig(rtr.CorsOrigins), rtr.Routes),
		cfg.ServerCfg)
}

func initAppConfig() app.Config {
	cfg := app.NewConfig()

	if err := cfg.Validate(); err != nil {
		log.Fatal("[initAppConfig] err: ", err)
	}

	return cfg
}

func registerProxies(ctx context.Context) {
	if err := proxy.Register(ctx, proxy.Config{
		LogServiceName:   true,
		IncludeQueryLogs: false,
		Name:             env.GetAndValidateF("URL_SHORTENER_SERVICE_NAME"),
		BaseURL:          env.GetAndValidateF("URL_SHORTENER_URL"),
	}); err != nil {
		log.Fatalf("register proxy failed: %v", err)
	}
}

func initRouter() handler.Router {
	return handler.Router{
		CorsOrigins: []string{"*"},
	}
}
