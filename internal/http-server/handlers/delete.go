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

type DeleteResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// NewDelete возвращает хендлер, удаляющий запись из таблицы
//
// @Summary Удалить запись о подписке
// @Description Удаляет запись по service_name и user_id
// @Tags subscriptions
// @Produce json
// @Param service_name path string true "Имя сервися"
// @Param user_id path string true "UUID пользователя"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/subscriptions/{service_name}/{user_id} [delete]
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
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("url param is empty"))
			return
		}

		if err := storage.Delete(serviceName, userID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Warn("record not found: %s, %s", serviceName, userID)
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, response.Error("record not found"))
				return
			}
			log.Error("Failed to delete record", slog.String("error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error"))
			return
		}

		log.Info("Record deleted successfully", slog.String("service_name", serviceName), slog.String("user_id", userID))
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, DeleteResponse{
			Status:  "success",
			Message: "record was deleted successfully",
		})
	}
}
