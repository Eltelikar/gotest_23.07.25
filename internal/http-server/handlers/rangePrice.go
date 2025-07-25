package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"gotest_23.07.25/internal/http-server/response"
)

type RangeRequestBody struct {
	StartDate   time.Time `json:"start_date" example:"2025-01-01T00:00:00Z"`
	EndDate     time.Time `json:"end_date" example:"2025-12-31T00:00:00Z"`
	ServiceName string    `json:"service_name" example:"Google"`
	UserID      string    `json:"user_id" example:"b1d4c0ec-9a4a-4e3a-9fdd-5e27d0be16fa"`
}

type RangeResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Price   uint64 `json:"price"`
}

type RangePrice interface {
	RangePrice(start_date time.Time, end_date time.Time, service_name string, user_id string) (uint64, error)
}

// NewRangePrice возвращает хендлер, возвращающий стоимость подписок в выбранном периоде
//
// @Summary Получить общую стоимость подписок за период
// @Description Подсчитывает общую стоимость подписок по start_date, end_date, service_name, user_id. service_name и user_id можно передать пустыми.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription_filter body RangeRequestBody true "фильтры для рассчета"
// @Success 200 {object} RangeResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/subscriptions/range-price [post]
func NewRangePrice(log *slog.Logger, storage RangePrice) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.NewRangePrice"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Info("RangePrice handler started")

		var rb RangeRequestBody
		if err := render.DecodeJSON(r.Body, &rb); err != nil {
			log.Error("Failed to decode request body", slog.String("error", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request body"))
			return
		}

		log.Debug("Decoded request body", slog.Any("request_body", rb))

		if rb.StartDate.IsZero() || rb.EndDate.IsZero() {
			log.Info("url param is empty")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("url param is empty"))
			return
		}

		if rb.StartDate.After(rb.EndDate) {
			log.Info("Start-date is after end-date")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("start date cannot be after end date"))
			return
		}

		ResPrice, err := storage.RangePrice(rb.StartDate, rb.EndDate, rb.ServiceName, rb.UserID)
		if err != nil {
			log.Error("Failed to get range price", slog.String("error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error"))
			return
		}
		log.Info("Get range price successfully", slog.Uint64("price", ResPrice))
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, RangeResponse{
			Status:  "success",
			Message: "Get range price successfully",
			Price:   ResPrice,
		})
	}

}
