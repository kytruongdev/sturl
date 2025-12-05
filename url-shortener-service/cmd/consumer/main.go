package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/config"
	shortUrlCtrl "github.com/kytruongdev/sturl/url-shortener-service/internal/controller/shorturl"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/app"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/db/pg"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/id"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/kafka"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository"
	redisRepo "github.com/kytruongdev/sturl/url-shortener-service/internal/repository/redis"
	"github.com/redis/go-redis/v9"
)

func main() {
	rootCtx := context.Background()

	globalCfg := config.NewGlobalConfig()
	if err := globalCfg.Validate(); err != nil {
		log.Fatal("invalid config: ", err)
	}

	if err := id.Init(1); err != nil {
		panic(err)
	}

	shutdown, err := monitoring.Init(rootCtx, globalCfg.MonitoringCfg)
	if err != nil {
		panic(err)
	}
	defer shutdown(rootCtx)

	// --- Setup db
	conn := initDB(globalCfg)
	defer conn.Close()

	kafkaProducer := kafka.NewProducer(globalCfg.KafkaCfg)
	defer kafkaProducer.Close()

	// --- Setup redis
	redisClient := initRedis(rootCtx, globalCfg)

	repo := repository.New(conn, redisClient)
	shortURLCtrl := shortUrlCtrl.New(repo)

	consumer := New(globalCfg.KafkaCfg, shortURLCtrl, kafkaProducer)

	appRunner := app.Runner{Name: globalCfg.KafkaCfg.ClientID + "-consumer"}
	if err := appRunner.Run(rootCtx, runner{consumer: consumer}); err != nil {
		monitoring.Log(rootCtx).Error().Err(err).Msg("consumer exited with error")
	}
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

type runner struct {
	consumer Consumer
}

func (r runner) Run(ctx context.Context) error {
	return r.consumer.Start(ctx)
}

func (r runner) Shutdown(ctx context.Context) error {
	return r.consumer.Stop(ctx)
}
