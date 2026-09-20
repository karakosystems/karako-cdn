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

	"github.com/karakosystems/karako-cdn/internal/server"
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	cfg := server.Config{
		BaseFQDN: os.Getenv("BASE_FQDN"),
		CDNFQDN:  os.Getenv("CDN_FQDN"),
		Addr:     ":" + envOr("PORT", "80"),
	}
	if cfg.BaseFQDN == "" {
		log.Fatal("BASE_FQDN is required (redirect target, e.g. example.com)")
	}
	assetsDir := envOr("ASSETS_DIR", "assets")

	srv, err := server.New(cfg, os.DirFS(assetsDir))
	if err != nil {
		log.Fatalf("loading assets from %s: %v", assetsDir, err)
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
