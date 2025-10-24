package main

import (
	"context"
	"database/sql"
	"log"

	shortUrlCtrl "github.com/kytruongdev/sturl/url-shortener-service/internal/controller/shorturl"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/handler"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/app"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/db/pg"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/httpserver"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/logger"
	redisRepo "github.com/kytruongdev/sturl/url-shortener-service/internal/repository/redis"
	shortUrlRepo "github.com/kytruongdev/sturl/url-shortener-service/internal/repository/shorturl"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := initAppConfig()

	// --- Setup logger
	rootLog := logger.New(cfg.ServerCfg.ServiceName, cfg.ServerCfg.LogLevel, cfg.ServerCfg.AppEnv)
	rootCtx := logger.ToContext(context.Background(), rootLog)
	l := logger.FromContext(rootCtx)

	l.Info().Msg("Starting app initialization")

	// --- Setup dependencies
	redisClient := initRedis(rootCtx, cfg)
	conn := initDB(cfg)
	defer conn.Close()

	shortURLRepo := shortUrlRepo.New(conn, redisClient)
	shortURLCtrl := shortUrlCtrl.New(shortURLRepo)
	rtr := initRouter(rootCtx, shortURLCtrl)

	l.Info().Msg("App initialization completed")

	// --- Start server
	httpserver.Start(httpserver.Handler(rootCtx, httpserver.NewCORSConfig(rtr.CorsOrigins), rtr.Routes),
		cfg.ServerCfg)
}

func initDB(cfg app.Config) *sql.DB {
	conn, err := pg.Connect(cfg.PGCfg.PGUrl)
	if err != nil {
		log.Fatal("[pg.Connect] err]: ", err)
	}

	return conn
}

func initAppConfig() app.Config {
	cfg := app.NewConfig()

	if err := cfg.Validate(); err != nil {
		log.Fatal("[initAppConfig] err: ", err)
	}

	return cfg
}

func initRedis(ctx context.Context, cfg app.Config) redisRepo.RedisClient {
	redisClient, err := redisRepo.NewRedisClient(ctx, &redis.Options{
		Addr: cfg.ServerCfg.RedisAddr,
		DB:   0,
	})

	if err != nil {
		log.Fatal("[initRedis] err: ", err)
	}

	return redisClient
}

func initRouter(ctx context.Context, shortURLCtrl shortUrlCtrl.Controller) handler.Router {
	return handler.Router{
		Ctx:          ctx,
		CorsOrigins:  []string{"*"},
		ShortURLCtrl: shortURLCtrl,
	}
}
