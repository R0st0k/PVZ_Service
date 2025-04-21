package handler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	api "pvz-service/api/generated"
	e "pvz-service/internal/errors"
	"pvz-service/internal/logger/sl"
	"pvz-service/internal/models"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func (h *Handler) CreatePVZ() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.CreatePVZ"

		log := h.log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req api.PostPvzJSONRequestBody

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

		pvz := models.PVZ{
			CityName: string(req.City),
		}

		if req.RegistrationDate == nil {
			pvz.RegistrationDate = time.Now()
		} else {
			pvz.RegistrationDate = *req.RegistrationDate
		}

		if req.Id == nil {
			pvz.ID = uuid.New()
		} else {
			pvz.ID = *req.Id
		}

		resp, err := h.pvzService.CreatePVZ(r.Context(), &pvz)
		if err != nil {
			log.Error("failed to create pvz", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, api.Error{Message: "failed to create pvz"})

			return
		}

		w.WriteHeader(http.StatusCreated)
		render.JSON(w, r, resp)
	}
}
