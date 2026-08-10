// Command karako-cdn is the CDN server.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	karakocdn "github.com/karakosystems/karako-cdn"
	"github.com/karakosystems/karako-cdn/internal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	cfg := server.Config{BaseFQDN: os.Getenv("BASE_FQDN"), Addr: ":" + port}
	if cfg.BaseFQDN == "" {
		cfg.BaseFQDN = "karakosystems.com"
	}

	srv, err := server.New(cfg, karakocdn.Assets)
	if err != nil {
		log.Fatal("loading embedded assets: ", err)
	}

	paths := srv.Paths()
	log.Printf("loaded %d resources: %v", len(paths), paths)
	log.Printf("redirect domain: %s", cfg.BaseFQDN)
	log.Printf("starting server on %s", cfg.Addr)

	httpSrv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           srv,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() { errCh <- httpSrv.ListenAndServe() }()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	case <-ctx.Done():
		log.Print("shutdown signal received, draining connections")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpSrv.Shutdown(shutdownCtx); err != nil {
			log.Print("shutdown: ", err)
		}
	}
}
