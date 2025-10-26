package main

import (
	"log/slog"
	"os"

	"github.com/gin-contrib/logger"
	// "github.com/gin-contrib/recovery"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"

	// "github.com/gin-gonic/gin/binding"
	// "github.com/gin-contrib/auth"

	"github.com/fatykhovar/url-shortner/internal/config"
	"github.com/fatykhovar/url-shortner/internal/http-server/handlers/redirect"
	"github.com/fatykhovar/url-shortner/internal/http-server/handlers/url/save"
	"github.com/fatykhovar/url-shortner/internal/lib/logger/sl"
	"github.com/fatykhovar/url-shortner/internal/storage/postgres"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("Logger initialized", slog.String("env", cfg.Env))
	log.Debug("Debug messag	e")

	storage, err := postgres.New(cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to initialize storage", sl.Err(err))
		os.Exit(1)
	}

	router := gin.Default()

	router.Use(requestid.New())    // добавляем X-Request-ID
	router.Use(logger.SetLogger()) // лог запросов
	// router.Use(mwLogger.New(log))  // твой кастомный логгер
	router.Use(gin.Recovery()) // обработка паник

	authorized := router.Group("/url", gin.BasicAuth(gin.Accounts{
		cfg.HTTPServer.User: cfg.HTTPServer.Password,
	}))

	authorized.POST("/", save.New(log, storage))
	router.GET("/:alias", redirect.New(log, storage))

	log.Info("Starting HTTP server", slog.String("address", cfg.HTTPServer.Address))

	if err := router.Run(cfg.HTTPServer.Address); err != nil {
		log.Error("failed to start HTTP server", sl.Err(err))
	}

	storage.Close()

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default: // If env config is invalid, set prod settings by default due to security
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
