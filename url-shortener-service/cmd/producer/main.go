package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/config"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/controller/outgoingevent"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/app"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/db/pg"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/id"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/kafka"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository"
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

	kafkaProducer := kafka.NewProducer(globalCfg.KafkaCfg)
	defer kafkaProducer.Close()

	producer := New(
		outgoingevent.New(repository.New(conn, nil), kafkaProducer),
		initProducerConfig())

	// --- Start Producer
	r := app.Runner{Name: globalCfg.KafkaCfg.ClientID}
	if err = r.Start(rootCtx, runner{producer}); err != nil {
		monitoring.Log(rootCtx).Error().Err(err).Msg("outbox worker exited with error")
	}
}

func loadGlobalConfig() config.GlobalConfig {
	cfg := config.NewGlobalConfig()

	if err := cfg.Validate(); err != nil {
		log.Fatal("[loadGlobalConfig] err: ", err)
	}

	return cfg
}

func initProducerConfig() ProducerConfig {
	pim, err := strconv.Atoi(os.Getenv("POLLING_INTERVAL_MS"))
	if err != nil {
		panic(err)
	}

	bs, err := strconv.Atoi(os.Getenv("BATCH_SIZE"))
	if err != nil {
		panic(err)
	}

	mr, err := strconv.Atoi(os.Getenv("MAX_RETRY"))
	if err != nil {
		panic(err)
	}

	return ProducerConfig{
		pollingInterval: time.Duration(pim) * time.Millisecond,
		batchSize:       bs,
		maxRetry:        mr,
	}
}

func initMonitoring(ctx context.Context, cfg monitoring.Config) (func(context.Context) error, error) {
	shutdown, err := monitoring.Init(ctx, monitoring.Config{
		ServiceName:     cfg.ServiceName,
		Env:             cfg.Env,
		OTLPEndpointURL: cfg.OTLPEndpointURL,
		LogPretty:       true,
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

// runner is an adapter to make the Producer worker implement app.Service
type runner struct {
	producer Producer
}

func (r runner) Run(ctx context.Context) error {
	// Producer.Start(ctx) is assumed to be an infinite loop that breaks when ctx.Done.
	return r.producer.Start(ctx)
}

// Shutdown here can be a no-op because ctx cancellation will cause Start to return.
func (r runner) Shutdown(_ context.Context) error {
	// If you want to add custom stop logic later, implement it here.
	return nil
}
