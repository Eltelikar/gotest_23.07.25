package handlers

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"gotest_23.07.25/internal/http-server/response"
)

type Delete interface {
	Delete(serviceName, userID string) error
}

func NewDelete(log *slog.Logger, storage Delete) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.NewDelete"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Info("Delete handler started")

		serviceName := chi.URLParam(r, "service_name")
		userID := chi.URLParam(r, "user_id")
		if serviceName == "" || userID == "" {
			log.Info("url param is empty")
			render.JSON(w, r, response.Error("url param is empty"))
			return
		}

		if err := storage.Delete(serviceName, userID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Warn("record not found: %s, %s", serviceName, userID)
				render.JSON(w, r, response.Error("record not found"))
				return
			}
			log.Error("Failed to delete record", slog.String("error", err.Error()))
			render.JSON(w, r, response.Error("internal error"))
			return
		}

		log.Info("Record deleted successfully", slog.String("service_name", serviceName), slog.String("user_id", userID))
		render.JSON(w, r, response.OK("Record deleted", nil))
	}
}
