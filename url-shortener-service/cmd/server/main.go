package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/config"
	shortUrlCtrl "github.com/kytruongdev/sturl/url-shortener-service/internal/controller/shorturl"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/handler"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/db/pg"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/httpserver"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	redisRepo "github.com/kytruongdev/sturl/url-shortener-service/internal/repository/redis"
	shortUrlRepo "github.com/kytruongdev/sturl/url-shortener-service/internal/repository/shorturl"
	"github.com/redis/go-redis/v9"
	"github.com/volatiletech/sqlboiler/boil"
)

func main() {
	rootCtx := context.Background()

	l := monitoring.Log(rootCtx)
	l.Info().Msg("Starting app initialization")

	// --- Load global config
	globalCfg := loadGlobalConfig()

	// --- Setup logging monitoring
	shutdown, err := initMonitoring(rootCtx, globalCfg.MonitoringCfg)
	if err != nil {
		panic(err)
	}
	defer shutdown(rootCtx)

	// --- Setup db
	conn := initDB(globalCfg)
	defer conn.Close()

	// --- Setup redis
	redisClient := initRedis(rootCtx, globalCfg)

	// --- Setup routers
	rtr := initRouter(conn, redisClient)

	l.Info().Msg("App initialization completed")

	// --- Start server
	httpserver.Start(httpserver.Handler(httpserver.NewCORSConfig(rtr.CorsOrigins), globalCfg.TransportMetaCfg, rtr.Routes),
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

func initDB(cfg config.GlobalConfig) *sql.DB {
	conn, err := pg.Connect(cfg.PGCfg.PGUrl)
	if err != nil {
		log.Fatal("[pg.Connect] err]: ", err)
	}

	return conn
}

func initRedis(ctx context.Context, cfg config.GlobalConfig) redisRepo.RedisClient {
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
