package handler

import (
	"log/slog"
	"net/http"
	api "pvz-service/api/generated"
	e "pvz-service/internal/errors"
	"pvz-service/internal/logger/sl"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

func (h *Handler) DeleteLastProduct() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.DeleteLastProduct"

		log := h.log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		pvzId := chi.URLParam(r, "pvzId")
		if pvzId == "" {
			log.Error("url param is empty")

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, api.Error{Message: "empty request"})

			return
		}

		log.Info("url param decoded", slog.Any("param", pvzId))

		id, err := uuid.Parse(pvzId)
		if err != nil {
			log.Error("invalid param", sl.Err(err))

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, api.Error{Message: "invalid param"})

			return
		}

		err = h.pvzService.DeleteLastProduct(r.Context(), id)
		if err == e.ErrNoActiveReception() {
			log.Error("no active reception", sl.Err(err))

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, api.Error{Message: "no active reception"})

			return
		}
		if err == e.ErrNoProduct() {
			log.Error("no products to delete", sl.Err(err))

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, api.Error{Message: "no products to delete"})

			return
		}
		if err != nil {
			log.Error("failed to delete product", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, api.Error{Message: "failed to delete product"})

			return
		}

		w.WriteHeader(http.StatusOK)
		render.NoContent(w, r)
	}
}
