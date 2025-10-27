package main

import (
	"context"
	"log"
	"time"

	"github.com/kytruongdev/sturl/api-gateway/internal/handler"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/app"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/env"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/httpserver"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/logger"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/proxy"
)

func main() {
	cfg := initAppConfig()

	// --- Setup logger
	rootLog := logger.New(cfg.ServerCfg.ServiceName, cfg.ServerCfg.LogLevel, cfg.ServerCfg.AppEnv)
	rootCtx := logger.ToContext(context.Background(), rootLog)
	l := logger.FromContext(rootCtx)

	l.Info().Msg("Starting app initialization")

	// --- Setup dependencies
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

func initRouter() handler.Router {
	return handler.Router{
		CorsOrigins: []string{"*"},
	}
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
