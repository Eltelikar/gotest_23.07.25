package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"gotest_23.07.25/internal/http-server/response"
	"gotest_23.07.25/internal/postgre"
)

type List interface {
	List() ([]postgre.RequestFields, error)
}

func NewList(log *slog.Logger, storage List) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.NewList"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Info("List handler started")

		subscriptions, err := storage.List()
		if err != nil {
			log.Error("Failed to list subscriptions", slog.String("error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error"))
			return
		}

		log.Info("Subscriptions listed successfully")
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, map[string]any{
			"status":        "success",
			"message":       "Subscriptions listed successfully",
			"subscriptions": subscriptions,
		})
	}
}
