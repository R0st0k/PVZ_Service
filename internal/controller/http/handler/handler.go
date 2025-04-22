package handler

import (
	"log/slog"
	"pvz-service/internal/metrics"
	"pvz-service/internal/service"
)

type Handler struct {
	log         *slog.Logger
	metrics     *metrics.Metrics
	authService service.AuthServiceInterface
	pvzService  service.PVZServiceInterface
}

func NewHandler(
	log *slog.Logger,
	metrics *metrics.Metrics,
	authService service.AuthServiceInterface,
	pvzService service.PVZServiceInterface,
) *Handler {
	return &Handler{
		log:         log,
		metrics:     metrics,
		authService: authService,
		pvzService:  pvzService,
	}
}
