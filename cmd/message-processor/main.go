package main

import (
	"context"
	"database/sql"
	"errors"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"insider-go-project/internal/message-processor/config"
	"insider-go-project/internal/message-processor/services/message-processor"
	"insider-go-project/internal/message-processor/storage/repository"
	httpapp "insider-go-project/internal/message-processor/transport/httpapp/v1/message-processor"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("loading configs", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		logger.Error("open DB", "err", err)
		return
	}

	repo := repository.NewMessagesRepository(db)
	logger.Info("Redis addr", "addr", cfg.RedisAddr)

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})

	service := messageprocessor.NewService(
		repo,
		logger,
		redisClient,
		cfg.WebhookURL,
		cfg.AuthKey,
		cfg.BatchSize,
	)

	e := echo.New()

	httpapp.NewHandler(service).RegisterRoutes(e)

	go func() {
		logger.Info("starting server", "addr", cfg.HTTPAddr)
		if err := e.Start(cfg.HTTPAddr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", "error", err)
	}
}
