package handler

import (
	"log/slog"
	"net/http"
	api "pvz-service/api/generated"
	"pvz-service/internal/logger/sl"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func (h *Handler) GetPVZsWithReceptions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.GetPVZsWithReceptions"

		log := h.log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		query := r.URL.Query()
		var (
			param     string
			err       error
			layout    string = time.RFC3339
			startDate time.Time
			endDate   time.Time
			page      int
			limit     int
		)

		param = query.Get("startDate")
		if param == "" {
			startDate = time.Time{}
		} else {
			startDate, err = time.Parse(layout, param)
			if err != nil {
				log.Error("invalid startDate param", sl.Err(err))

				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, api.Error{Message: "invalid startDate param"})

				return
			}
		}

		param = query.Get("endDate")
		if param == "" {
			endDate = time.Now()
		} else {
			endDate, err = time.Parse(layout, param)
			if err != nil {
				log.Error("invalid endDate param", sl.Err(err))

				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, api.Error{Message: "invalid endDate param"})

				return
			}
		}

		param = query.Get("page")
		if param == "" {
			page = 1
		} else {
			page, err = strconv.Atoi(param)
			if err != nil || page < 1 {
				log.Error("invalid page param", sl.Err(err))

				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, api.Error{Message: "invalid page param"})

				return
			}
		}

		param = query.Get("limit")
		if param == "" {
			limit = 10
		} else {
			limit, err = strconv.Atoi(param)
			if err != nil || limit < 1 || limit > 30 {
				log.Error("invalid limit param", sl.Err(err))

				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, api.Error{Message: "invalid limit param"})

				return
			}
		}

		log.Info("query param decoded and validated",
			slog.Any("startDate", startDate),
			slog.Any("endDate", endDate),
			slog.Any("page", page),
			slog.Any("limit", limit),
		)

		resp, err := h.pvzService.GetPVZsWithReceptions(r.Context(), startDate, endDate, page, limit)
		if err != nil {
			log.Error("failed to get pvz list", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, api.Error{Message: "failed to get pvz list"})

			return
		}

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, resp)
	}
}
