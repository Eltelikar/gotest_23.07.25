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

type Read interface {
	Read(service_name, user_id string) (*postgre.RequestFields, error)
}

// NewRead возвращает хендлер, возвращающий информацию о выбранной подписке
//
// @Summary Получить информацию о подписке
// @Description Возвращает информацию о подписке по service_name и user_id
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param service_name path string true "Имя сервиса"
// @Param user_id path string true "UUID пользователя"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/subscriptions/{service_name}/{user_id} [get]
func NewRead(log *slog.Logger, storage Read) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.NewRead"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Info("Read handler started")

		serviceName := chi.URLParam(r, "service_name")
		userID := chi.URLParam(r, "user_id")

		if serviceName == "" || userID == "" {
			log.Info("url param is empty")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("url param is empty"))
			return
		}

		rb, err := storage.Read(serviceName, userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Warn("record not found: %s, %s", serviceName, userID)
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, response.Error("record not found"))
				return
			}
			log.Error("Failed to read record", slog.String("error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error"))
			return
		}

		log.Info("Record read successfully", slog.Any("record", rb))
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, response.OK("Record read successfully", rb))
	}

}
