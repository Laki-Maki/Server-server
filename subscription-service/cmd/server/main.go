package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"subscription-service/internal/config"
	"subscription-service/internal/db"
	"subscription-service/internal/handler"
	"subscription-service/internal/logger"
	"subscription-service/internal/service"
)

func main() {
	// загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}
	log := logger.New(cfg.LogLevel)

	if cfg.DatabaseURL == "" {
		log.Fatal().Msg("database url is empty; set DATABASE_URL or DB_* variables")
	}
	if cfg.Port == "" {
		log.Fatal().Msg("port is empty; set PORT in config")
	}

	// подключение к БД
	dbConn, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("db connect failed")
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			log.Error().Err(err).Msg("db close error")
		}
	}()

	// инициализация зависимостей
	repo := db.NewStore(dbConn)
	svc := service.New(repo, log)
	router := handler.NewRouter(handler.NewHandler(svc, log), log)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.ServerReadTimeout,
		WriteTimeout: cfg.ServerWriteTimeout,
	}

	// graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info().
			Str("addr", srv.Addr).
			Msg("server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server listen failed")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("termination signal received, shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server shutdown failed")
	} else {
		log.Info().Msg("server shutdown completed")
	}
}
