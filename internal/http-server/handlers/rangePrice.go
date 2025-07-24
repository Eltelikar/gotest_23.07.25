package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"gotest_23.07.25/internal/http-server/response"
)

type RequestBody struct {
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	ServiceName string    `json:"service_name"`
	UserID      string    `json:"user_id"`
}

type RangePrice interface {
	RangePrice(start_date time.Time, end_date time.Time, service_name string, user_id string) (uint64, error)
}

func NewRangePrice(log *slog.Logger, storage RangePrice) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.NewRangePrice"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Info("RangePrice handler started")

		var rb RequestBody
		if err := render.DecodeJSON(r.Body, &rb); err != nil {
			log.Error("Failed to decode request body", slog.String("error", err.Error()))
			render.JSON(w, r, response.Error("invalid request body"))
			return
		}

		log.Debug("Decoded request body", slog.Any("request_body", rb))

		if rb.StartDate.IsZero() || rb.EndDate.IsZero() || rb.ServiceName == "" || rb.UserID == "" {
			log.Info("url param is empty")
			render.JSON(w, r, response.Error("url param is empty"))
			return
		}

		if rb.StartDate.After(rb.EndDate) {
			log.Info("Start-date is after end-date")
			render.JSON(w, r, response.Error("start date cannot be after end date"))
			return
		}

		price, err := storage.RangePrice(rb.StartDate, rb.EndDate, rb.ServiceName, rb.UserID)
		if err != nil {
			log.Error("Failed to get range price", slog.String("error", err.Error()))
			render.JSON(w, r, response.Error("internal error"))
			return
		}
		log.Info("Get range price successfully", slog.Uint64("price", price))
		render.JSON(w, r, map[string]any{
			"status":  "success",
			"message": "Get range price successfully",
			"price":   price,
		})
	}

}
