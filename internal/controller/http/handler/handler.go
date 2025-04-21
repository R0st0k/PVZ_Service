package handler

import (
	"log/slog"
	"pvz-service/internal/service"
)

type Handler struct {
	log         *slog.Logger
	authService service.AuthServiceInterface
	pvzService  service.PVZServiceInterface
}

func NewHandler(
	log *slog.Logger,
	authService service.AuthServiceInterface,
	pvzService service.PVZServiceInterface,
) *Handler {
	return &Handler{
		log:         log,
		authService: authService,
		pvzService:  pvzService,
	}
}
