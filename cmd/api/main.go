package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/MonicaMell/task-api/internal/config"
	"github.com/MonicaMell/task-api/internal/repository"
	"github.com/MonicaMell/task-api/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type application struct {
	config *config.Config
	logger *slog.Logger
	db     *pgxpool.Pool
	tasks  *service.TaskService
	auth   *service.AuthService
}

func main() {
	if err := run(); err != nil {
		slog.Error("startup failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	_ = godotenv.Load()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := openDB(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer db.Close()
	logger.Info("database connection pool established")

	taskRepo := repository.NewTaskRepository(db)
	taskService := service.NewTaskService(taskRepo)

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo)

	app := &application{
		config: cfg,
		logger: logger,
		db:     db,
		tasks:  taskService,
		auth:   authService,
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Addr,
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Info("starting server", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func openDB(dsn string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
