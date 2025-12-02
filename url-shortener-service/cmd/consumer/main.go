package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/config"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/app"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/db/pg"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/id"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository"
)

func main() {
	ctx := context.Background()

	globalCfg := config.NewGlobalConfig()
	if err := globalCfg.Validate(); err != nil {
		log.Fatal("invalid config: ", err)
	}

	if err := id.Init(1); err != nil {
		panic(err)
	}

	shutdown, err := monitoring.Init(ctx, globalCfg.MonitoringCfg)
	if err != nil {
		panic(err)
	}
	defer shutdown(ctx)

	// --- Setup db
	conn := initDB(globalCfg)
	defer conn.Close()

	consumer := New(globalCfg.KafkaCfg, repository.New(conn, nil))

	appRunner := app.Runner{Name: globalCfg.KafkaCfg.ClientID + "-consumer"}
	if err := appRunner.Run(ctx, runner{consumer: consumer}); err != nil {
		monitoring.Log(ctx).Error().Err(err).Msg("consumer exited with error")
	}
}

func initDB(cfg config.GlobalConfig) *sql.DB {
	conn, err := pg.Connect(cfg.PGCfg.PGUrl)
	if err != nil {
		log.Fatal("[pg.Connect] err]: ", err)
	}

	return conn
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
