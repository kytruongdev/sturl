package main

import (
	"context"
	"log"
	"time"

	"github.com/kytruongdev/sturl/api-gateway/internal/handler"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/app"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/env"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/httpserver"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/monitoring/logging"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/monitoring/tracing"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/proxy"
)

func main() {
	cfg := initAppConfig()

	// --- Setup logging
	rootLog := logging.New(logging.FromEnv())
	rootCtx := logging.ToContext(context.Background(), rootLog)
	l := logging.FromContext(rootCtx)

	// --- Setup tracing
	shutdown, err := tracing.Init(rootCtx, tracing.FromEnv())
	if err != nil {
		l.Fatal().Err(err).Msg("failed to initialize tracing")
	}
	defer shutdown(context.Background())

	l.Info().Msg("Starting app initialization")

	// --- Setup proxies
	registerProxies()

	rtr := initRouter()

	l.Info().Msg("App initialization completed")

	// --- Start server
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

func registerProxies() {
	common := proxy.ServiceConfig{
		ResponseTimeout:  5 * time.Second,
		IdleConnTimeout:  30 * time.Second,
		MaxIdleConns:     50,
		LogServiceName:   true,
		IncludeQueryLogs: false,
	}

	shortenerCfg := common
	shortenerCfg.Name = env.GetAndValidateF("URL_SHORTENER_SERVICE_NAME")
	shortenerCfg.BaseURL = env.GetAndValidateF("URL_SHORTENER_URL")

	if err := proxy.Register(shortenerCfg); err != nil {
		log.Fatalf("register proxy failed: %v", err)
	}

	// The comment-out code block below is an example of how to register another proxy
	//userCfg := common
	//userCfg.Name = "user-service"
	//userCfg.BaseURL = env.GetAndValidateF("URL_USER_SERVICE")
	//userCfg.ResponseTimeout = 3 * time.Second // everride config example

	//if err := proxy.Register(userCfg); err != nil {
	//	log.Fatalf("register proxy failed: %v", err)
	//}
}

func initRouter() handler.Router {
	return handler.Router{
		CorsOrigins: []string{"*"},
	}
}
