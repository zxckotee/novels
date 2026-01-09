package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"novels-backend/internal/config"
	"novels-backend/internal/database"
	"novels-backend/internal/http/routes"
	"novels-backend/pkg/logger"

	"github.com/rs/zerolog"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.Load()

	// Инициализируем логгер
	log := logger.New(cfg.IsDevelopment())

	log.Info().
		Str("env", cfg.Server.Env).
		Str("port", cfg.Server.Port).
		Msg("Starting Novels API server")

	// Подключаемся к базе данных
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	log.Info().Msg("Connected to database")

	// Запускаем миграции
	if err := database.RunMigrations(db); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}

	log.Info().Msg("Database migrations completed")

	// Создаём роутер
	router := routes.NewRouter(db, cfg, log)

	// Создаём HTTP сервер
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Запускаем сервер в горутине
	go func() {
		log.Info().Msgf("Server is listening on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited gracefully")
}

// logger package
func init() {
	zerolog.TimeFieldFormat = time.RFC3339
}
