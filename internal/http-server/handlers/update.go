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
	"gotest_23.07.25/internal/postgre"
)

type Update interface {
	Update(service_name, user_id string, rb postgre.RequestUpdateFields) error
}

// NewUpdate возвращает хендлер, изменяющий информацию о подписке
//
// @Summary Изменить информацию о подписке
// @Description Возвращает информацию о подписке по service_name и user_id
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param service_name path string true "Имя подписки изменяемой записи"
// @Param user_id path string true "UUID пользователя изменяемой записи"
// @Param newFields body postgre.RequestUpdateFields true "Новая информация о подписке"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/subscriptions/{service_name}/{user_id} [put]
func NewUpdate(log *slog.Logger, storage Update) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.NewUpdate"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Info("Update handler started")

		serviceName := chi.URLParam(r, "service_name")
		userID := chi.URLParam(r, "user_id")

		if serviceName == "" || userID == "" {
			log.Info("url param is empty")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("url param is empty"))
			return
		}

		var rb postgre.RequestUpdateFields

		if err := render.DecodeJSON(r.Body, &rb); err != nil {
			log.Error("Failed to decode request body", slog.String("error", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request body"))
			return
		}

		log.Debug("Decoded request body", slog.Any("request_body", rb))

		if err := storage.Update(serviceName, userID, rb); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Warn("record not found",
					slog.String("service_name", serviceName),
					slog.String("user_id", userID),
				)
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, response.Error("record not found"))
				return
			}
			log.Error("Failed to update record", slog.String("error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error"))
			return
		}

		log.Info("Record updated successfully", slog.Any("record", rb))
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, response.OKUpdate("Record updated successfully", &rb))
	}
}
