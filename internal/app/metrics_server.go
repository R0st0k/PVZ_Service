package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"pvz-service/internal/config"
	"pvz-service/internal/logger/sl"
	"pvz-service/internal/metrics"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartMetricsServer(cfg *config.Config, log *slog.Logger, metrics *metrics.Metrics) func(*sync.WaitGroup) {
	// Настройка маршрутизации
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Prometheus.Host, cfg.Prometheus.Port),
		Handler:      metricsMux,
		ReadTimeout:  cfg.Prometheus.Timeout,
		WriteTimeout: cfg.Prometheus.Timeout,
		IdleTimeout:  cfg.Prometheus.IdleTimeout,
	}

	// Starting server
	go func() {
		log.Info("starting metrics server", slog.String("address", fmt.Sprintf("%s:%s", cfg.Prometheus.Host, cfg.Prometheus.Port)))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("metrics server error", sl.Err(err))
		}
	}()

	// Graceful Stop
	return func(wg *sync.WaitGroup) {
		defer wg.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer shutdownCancel()

		// Shutdown server
		log.Info("Shutting down metrics server")
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("metrics server shutdown error", sl.Err(err))
		} else {
			log.Info("metrics server gracefully stopped")
		}
	}
}
