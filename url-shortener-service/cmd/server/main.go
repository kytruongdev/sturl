package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/config"
	shortUrlCtrl "github.com/kytruongdev/sturl/url-shortener-service/internal/controller/shorturl"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/handler"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/app"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/db/pg"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/httpserver"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/id"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository"
	redisRepo "github.com/kytruongdev/sturl/url-shortener-service/internal/repository/redis"
	"github.com/redis/go-redis/v9"
)

func main() {
	rootCtx := context.Background()

	// --- Load global config
	globalCfg := loadGlobalConfig()

	if err := id.Init(1); err != nil {
		panic(err)
	}

	// --- Setup monitoring
	shutdown, err := initMonitoring(rootCtx, globalCfg.MonitoringCfg)
	if err != nil {
		panic(err)
	}
	defer shutdown(rootCtx)

	// --- Setup db
	conn := initDB(globalCfg)
	defer conn.Close()

	l := monitoring.Log(rootCtx)

	// --- Setup redis
	redisClient := initRedis(rootCtx, globalCfg)

	// --- Setup routers
	rtr := initRouter(conn, redisClient)

	l.Info().Msgf("%v service started", globalCfg.ServerCfg.ServiceName)

	// --- Start server
	svcName := globalCfg.ServerCfg.ServiceName
	if err = app.New(svcName).Run(
		rootCtx,
		runner{s: initHTTPServer(globalCfg, rtr, redisClient, conn)}); err != nil {
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
		LogPretty:       os.Getenv("LOG_PRETTY") == "true",
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
	const (
		selectedDB   = 0
		dialTimeout  = 5 * time.Second
		readTimeout  = 3 * time.Second
		writeTimeout = 3 * time.Second
		poolSize     = 10
		minIdleConns = 5
		maxRetries   = 3
	)

	redisClient, err := redisRepo.NewRedisClient(ctx, &redis.Options{
		Addr:         cfg.ServerCfg.RedisAddr,
		DB:           selectedDB,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		MaxRetries:   maxRetries,
	})

	if err != nil {
		log.Fatal("[initRedis] err: ", err)
	}

	return redisClient
}

func initHTTPServer(globalCfg config.GlobalConfig, rtr handler.Router, redisClient redisRepo.RedisClient, conn *sql.DB) http.Server {
	const (
		readTimeout    = 10 * time.Second
		writeTimeout   = 10 * time.Second
		idleTimeout    = 120 * time.Second
		maxHeaderBytes = 1 << 20
	)

	return http.Server{
		Addr: globalCfg.ServerCfg.ServerAddr,
		Handler: httpserver.HandlerWithHealth(
			httpserver.NewCORSConfig(rtr.CorsOrigins),
			globalCfg.TransportMetaCfg,
			rtr.Routes,
			httpserver.ReadinessConfig{
				DB:    conn,
				Redis: redisClient,
			},
		),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		IdleTimeout:    idleTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}
}

func initRouter(conn *sql.DB, redisClient redisRepo.RedisClient) handler.Router {
	repo := repository.New(conn, redisClient)
	shortURLCtrl := shortUrlCtrl.New(repo)

	return handler.Router{
		CorsOrigins:  []string{"*"},
		ShortURLCtrl: shortURLCtrl,
	}
}

// runner is an adapter to make http.Server implement app.Service
type runner struct {
	s http.Server
}

func (h runner) Run(ctx context.Context) error {
	// ctx is not used directly here: Shutdown will cause ListenAndServe to unblock.
	if err := h.s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		monitoring.Log(ctx).Error().Err(err).Msg("url-shortener-service exited with error")
		return err
	}
	return nil
}

func (h runner) Shutdown(ctx context.Context) error {
	monitoring.Log(ctx).Warn().Msg("url-shortener-service exited")
	return h.s.Shutdown(ctx)
}
