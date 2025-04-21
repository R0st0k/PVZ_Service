package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"pvz-service/internal/config"
	"pvz-service/internal/logger/sl"
	"syscall"
	"time"
)

func StartHTTPServer(cfg *config.Config, log *slog.Logger, router *http.Handler) {
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler:      *router,
		ReadTimeout:  cfg.HTTP.Timeout,
		WriteTimeout: cfg.HTTP.Timeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	// Starting server
	go func() {
		log.Info("starting server", slog.String("address", fmt.Sprintf("%s:%s", cfg.HTTP.Host, cfg.HTTP.Port)))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http api server error", sl.Err(err))
			os.Exit(1)
		}
	}()

	// Terminate
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer shutdownCancel()

	// Shutdown server
	log.Info("Shutting down HTTP server")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP server shutdown error", sl.Err(err))
	} else {
		log.Info("HTTP server gracefully stopped")
	}
}
