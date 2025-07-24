package handlers

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"gotest_23.07.25/internal/http-server/response"
	pgr "gotest_23.07.25/internal/postgre"
)

type Update interface {
	Update(rb pgr.RequestFields) error
}

func NewUpdate(log *slog.Logger, storage Update) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.NewUpdate"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Info("Update handler started")

		var rb pgr.RequestFields

		if err := render.DecodeJSON(r.Body, &rb); err != nil {
			log.Error("Failed to decode request body", slog.String("error", err.Error()))
			render.JSON(w, r, response.Error("invalid request body"))
			return
		}

		log.Debug("Decoded request body", slog.Any("request_body", rb))

		if err := storage.Update(rb); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Warn("record not found: %s, %s", rb.ServiceName, rb.UserId)
				render.JSON(w, r, response.Error("record not found"))
				return
			}
			log.Error("Failed to update record", slog.String("error", err.Error()))
			render.JSON(w, r, response.Error("internal error"))
			return
		}

		log.Info("Record updated successfully", slog.Any("record", rb))
		render.JSON(w, r, response.OK("Record updated successfully", &rb))
	}
}
