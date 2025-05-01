package main

import (
	"context"
	"log"
	
	"github.com/kytruongdev/sturl/api-gateway/cmd/banner"
	"github.com/kytruongdev/sturl/api-gateway/internal/handler"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/app"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/httpserver"
)

func main() {
	banner.Print()

	ctx := context.Background()

	cfg, err := initAppConfig()
	if err != nil {
		log.Fatal(err)
	}

	rtr, err := initRouter(ctx)
	if err != nil {
		log.Fatal(err)
	}

	httpserver.Start(httpserver.Handler(httpserver.NewCORSConfig(rtr.CorsOrigins), rtr.Routes),
		cfg.ServerCfg)
}

func initAppConfig() (app.Config, error) {
	cfg := app.NewConfig()

	return cfg, cfg.Validate()
}

func initRouter(ctx context.Context) (handler.Router, error) {
	return handler.Router{
		Ctx:         ctx,
		CorsOrigins: []string{"*"},
	}, nil
}
