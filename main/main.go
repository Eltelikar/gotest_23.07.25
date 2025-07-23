package main

import (
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"gotest_23.07.25/internal/config"
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
	pathMakeNew = "/task"
	pathAddTask = "/task/{id}/"
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

	// TODO: make init logger func
	log := setupLogger(cfg.Env)
	slog.SetDefault(log)

	slog.Info("Starting service", slog.String("env", cfg.Env))
	slog.Debug("Debug messages are enabled")
	slog.Error("Error messages are enabled")

	//TODO: make init storage func
	storage, err := postgre.New(config.GetStorageLink(cfg))
	if err != nil {
		slog.Error("failed to init storage: %w", slog.String("error", err.Error()))
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	// TODO: middleware logger
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	// TODO: start service
}

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

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
