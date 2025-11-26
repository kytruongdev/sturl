package main

import (
	"context"
	"log"
	"net/http"

	"github.com/kytruongdev/sturl/api-gateway/internal/config"
	"github.com/kytruongdev/sturl/api-gateway/internal/handler"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/app"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/env"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/httpserver"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/proxy"
)

func main() {
	rootCtx := context.Background()

	// --- Load global config
	globalCfg := loadGlobalConfig()

	// --- Setup monitoring
	shutdown, err := initMonitoring(rootCtx, globalCfg.MonitoringCfg)
	if err != nil {
		panic(err)
	}
	defer shutdown(rootCtx)

	l := monitoring.Log(rootCtx)

	// --- Setup proxies
	registerProxies(rootCtx)

	// --- Setup routers
	rtr := initRouter()

	l.Info().Msgf("%v started", globalCfg.ServerCfg.ServiceName)

	// --- Start server
	svcName := globalCfg.ServerCfg.ServiceName
	r := app.Runner{Name: svcName}
	if err = r.Start(
		rootCtx,
		runner{
			s: &http.Server{
				Addr:    globalCfg.ServerCfg.ServerAddr,
				Handler: httpserver.Handler(httpserver.NewCORSConfig(rtr.CorsOrigins), rtr.Routes),
			}}); err != nil {
		l.Error().Err(err).Msgf("%v exited with error", svcName)
	}
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

// runner is an adapter to make http.Server implement app.Service
type runner struct {
	s *http.Server
}

func (h runner) Run(ctx context.Context) error {
	// ctx is not used directly here: Shutdown will cause ListenAndServe to unblock.
	if err := h.s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		monitoring.Log(ctx).Error().Err(err).Msg("api-gateway exited with error")
		return err
	}
	return nil
}

func (h runner) Shutdown(ctx context.Context) error {
	monitoring.Log(ctx).Warn().Msg("api-gateway exited")
	return h.s.Shutdown(ctx)
}
