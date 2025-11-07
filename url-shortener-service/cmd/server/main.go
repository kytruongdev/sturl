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
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	redisRepo "github.com/kytruongdev/sturl/url-shortener-service/internal/repository/redis"
	shortUrlRepo "github.com/kytruongdev/sturl/url-shortener-service/internal/repository/shorturl"
	"github.com/redis/go-redis/v9"
	"github.com/volatiletech/sqlboiler/boil"
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

	// --- Setup db
	conn := initDB(cfg)
	defer conn.Close()

	// --- Setup redis
	redisClient := initRedis(rootCtx, cfg)

	// --- Setup routers
	rtr := initRouter(conn, redisClient)

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

func initDB(cfg app.Config) *sql.DB {
	conn, err := pg.Connect(cfg.PGCfg.PGUrl)
	if err != nil {
		log.Fatal("[pg.Connect] err]: ", err)
	}

	return conn
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

func initRouter(conn boil.ContextExecutor, redisClient redisRepo.RedisClient) handler.Router {
	shortURLRepo := shortUrlRepo.New(conn, redisClient)
	shortURLCtrl := shortUrlCtrl.New(shortURLRepo)

	return handler.Router{
		CorsOrigins:  []string{"*"},
		ShortURLCtrl: shortURLCtrl,
	}
}
