package main

import (
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"gotest_23.07.25/internal/config"
	"gotest_23.07.25/internal/http-server/handlers"
	"gotest_23.07.25/internal/http-server/middlewares/logger"
	"gotest_23.07.25/internal/lib/slogpretty"
	"gotest_23.07.25/internal/postgre"
)

// logger levels:
const (
	envLocal = "local"
	envDebug = "debug"
	envProd  = "prod"
)

// api methods addresses:
const (
	createSubscription = "/api/v1/subscriptions"                          // post
	listSubscriptions  = "/api/v1/subscriptions"                          // get
	readSubscription   = "/api/v1/subscriptions/{service_name}/{user_id}" // get
	deleteSubscription = "/api/v1/subscriptions/{service_name}/{user_id}" // delete
	updateSubscription = "/api/v1/subscriptions/{service_name}/{user_id}" // put
	rangePrice         = "/api/v1/subscriptions/range-price"              // get

)

func main() {
	err := godotenv.Load("../config.env")
	if err != nil {
		err = godotenv.Load("config.env")
		if err != nil {
			slog.Error("failed to load .env file", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}

	cfg := config.MustLoad()
	slog.Info("Config file loaded successfully")

	log := setupLogger(cfg.Env)
	slog.SetDefault(log)
	slog.Info("Starting service", slog.String("env", cfg.Env))
	slog.Debug("Debug messages are enabled")
	slog.Error("Error messages are enabled")

	storage, err := initStorage(cfg)
	if err != nil {
		os.Exit(1)
	}

	router := initRouter(log)
	initHandlers(log, router, storage)

	// TODO: start server
}

// initStorage инициализирует in-memory хранилище и возвращает указатель на него
func initStorage(cfg *config.Config) (*postgre.Storage, error) {
	slog.Info("Init storage started")
	storage, err := postgre.New(config.GetStorageLink(cfg))
	if err != nil {
		slog.Error("failed to init storage: %w", slog.String("error", err.Error()))
		return nil, err
	}

	slog.Info("Init storage done successfully")
	return storage, nil
}

func initHandlers(log *slog.Logger, router *chi.Mux, storage *postgre.Storage) {
	slog.Info("Init handlers started")
	router.Post(createSubscription, handlers.NewCreate(log, storage))
	router.Get(listSubscriptions, handlers.NewList(log, storage))
	router.Get(readSubscription, handlers.NewRead(log, storage))
	router.Delete(deleteSubscription, handlers.NewDelete(log, storage))
	router.Put(updateSubscription, handlers.NewUpdate(log, storage))
	router.Get(rangePrice, handlers.NewRangePrice(log, storage))
	slog.Info("Handlers initialization successfully")
}

func initRouter(log *slog.Logger) *chi.Mux {
	slog.Info("Starting router")
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	slog.Debug("Middlewares used successfully",
		slog.String("middleware", "middleware/RequestID"),
		slog.String("middleware", "logger/New"),
		slog.String("middleware", "middleware/Recoverer"),
		slog.String("middleware", "middleware/URLFormat"),
	)
	slog.Info("Router and middlewares started")
	return router
}

// setupLogger настраивает вывод логгера в зависимости от настроек окружения
func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDebug:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}

// setupPrettySlog настраивает пакет PrettySlog
func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
