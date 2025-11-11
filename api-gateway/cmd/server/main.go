package main

import (
	"context"
	"log"

	"github.com/kytruongdev/sturl/api-gateway/internal/config"
	"github.com/kytruongdev/sturl/api-gateway/internal/handler"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/env"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/httpserver"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/proxy"
)

func main() {
	rootCtx := context.Background()

	log.Println("Starting app initialization")

	// --- Load global config
	globalCfg := loadGlobalConfig()

	// --- Setup logging monitoring
	shutdown, err := initMonitoring(rootCtx, globalCfg.MonitoringCfg)
	if err != nil {
		panic(err)
	}
	defer shutdown(rootCtx)

	// --- Setup proxies
	registerProxies(rootCtx)

	// --- Setup routers
	rtr := initRouter()

	log.Println("App initialization completed")

	// --- Start server
	httpserver.Start(httpserver.Handler(httpserver.NewCORSConfig(rtr.CorsOrigins), rtr.Routes),
		globalCfg.ServerCfg)
}

func loadGlobalConfig() config.GlobalConfig {
	cfg := config.NewGlobalConfig()

	if err := cfg.Validate(); err != nil {
		log.Fatal("[loadGlobalConfig] err: ", err)
	}

	return cfg
}

func initMonitoring(ctx context.Context, cfg monitoring.Config) (func(context.Context) error, error) {
	shutdown, err := monitoring.Init(ctx, monitoring.Config{
		ServiceName:     cfg.ServiceName,
		Env:             cfg.Env,
		LogPretty:       true,
		OTLPEndpointURL: cfg.OTLPEndpointURL,
	})

	if err != nil {
		log.Fatal("[initMonitoring] err: ", err)
	}

	return shutdown, nil
}

func registerProxies(ctx context.Context) {
	if err := proxy.Register(ctx, proxy.Config{
		UpstreamServiceName:    env.GetAndValidateF("URL_SHORTENER_SERVICE_NAME"),
		UpstreamServiceBaseURL: env.GetAndValidateF("URL_SHORTENER_URL"),
	}); err != nil {
		log.Fatalf("register proxy failed: %v", err)
	}
}

func initRouter() handler.Router {
	return handler.Router{
		CorsOrigins: []string{"*"},
	}
}
