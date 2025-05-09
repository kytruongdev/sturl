package main

import (
	"context"
	"log"

	"github.com/kytruongdev/sturl/url-shortener-service/cmd/banner"
	shortUrlCtrl "github.com/kytruongdev/sturl/url-shortener-service/internal/controller/shorturl"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/handler"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/app"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/db/pg"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/httpserver"
	shortUrlRepo "github.com/kytruongdev/sturl/url-shortener-service/internal/repository/shorturl"
	"github.com/redis/go-redis/v9"
)

func main() {
	banner.Print()

	cfg, err := initAppConfig()
	if err != nil {
		log.Fatal(err)
	}

	conn, err := pg.Connect(cfg.PGCfg.PGUrl)
	if err != nil {
		log.Fatal("[PG connection error] ", err)
	}
	defer conn.Close()

	ctx := context.Background()

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.ServerCfg.RedisAddr,
	})

	shortURLRepo := shortUrlRepo.New(conn, redisClient)
	shortURLCtrl := shortUrlCtrl.New(shortURLRepo)

	rtr, err := initRouter(ctx, shortURLCtrl)
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

func initRouter(ctx context.Context, shortURLCtrl shortUrlCtrl.Controller) (handler.Router, error) {
	return handler.Router{
		Ctx:          ctx,
		CorsOrigins:  []string{"*"},
		ShortURLCtrl: shortURLCtrl,
	}, nil
}
