package httpserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Start launches the HTTP server with graceful shutdown support
func Start(handler http.Handler, cfg Config) {
	svr := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: handler,
	}

	go func() {
		svr.ListenAndServe()
	}()

	defer stop(svr)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(fmt.Sprint(<-ch))
}

func stop(svr *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := svr.Shutdown(ctx); err != nil {
		log.Printf("Could not shut down server correctly: %v\n", err)
		os.Exit(1)
	}
}
