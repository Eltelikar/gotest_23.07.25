package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"gotest_23.07.25/internal/http-server/response"
	pgr "gotest_23.07.25/internal/postgre"
)

type Create interface {
	Create(rb pgr.RequestFields) (string, error)
}

func NewCreate(log *slog.Logger, storage Create) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.NewCreate"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Info("Create handler started")

		var rb pgr.RequestFields

		if err := render.DecodeJSON(r.Body, &rb); err != nil {
			log.Error("Failed to decode request body", slog.String("error", err.Error()))
			render.JSON(w, r, response.Error("invalid request body"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug("Decoded request body", slog.Any("request_body", rb))

		_, err := storage.Create(rb)
		if err != nil {
			log.Error("Failed to create record", slog.String("error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error"))
			return
		}
		log.Info("New record created successfully", slog.Any("record", rb))
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, response.OK("New record created", &rb))
	}
}
