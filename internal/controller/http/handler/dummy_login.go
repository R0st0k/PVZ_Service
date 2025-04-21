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

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

func (h *Handler) DummyLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.DummyLogin"

		log := h.log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req api.PostDummyLoginJSONRequestBody

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

		token, err := h.authService.DummyLogin(models.UserRole(req.Role))
		if err != nil {
			log.Error("failed to create token", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, api.Error{Message: "failed to create token"})

			return
		}

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, api.Token(token))
	}
}
