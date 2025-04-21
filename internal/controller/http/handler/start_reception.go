package handler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	api "pvz-service/api/generated"
	e "pvz-service/internal/errors"
	"pvz-service/internal/logger/sl"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

func (h *Handler) StartReception() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.StartReception"

		log := h.log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req api.PostReceptionsJSONRequestBody

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, api.Error{Message: "empty request"})

			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, api.Error{Message: "failed to decode request"})

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, api.Error{Message: e.ValidationError(validateErr)})

			return
		}

		reception, err := h.pvzService.StartReception(r.Context(), req.PvzId)
		if err == e.ErrCityNotAllowed() {
			log.Error("city not allowed", sl.Err(err))

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, api.Error{Message: "city not allowed"})

			return
		}
		if err == e.ErrActiveReceptionExists() {
			log.Error("active reception exists", sl.Err(err))

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, api.Error{Message: "active reception exists"})

			return
		}
		if err != nil {
			log.Error("failed to start reception", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, api.Error{Message: "failed to start reception"})

			return
		}

		w.WriteHeader(http.StatusCreated)
		render.JSON(w, r, reception)
	}
}
