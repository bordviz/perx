package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"perx/internal/config"
	"perx/internal/handlers"
	"perx/internal/lib/logger/sl"
	"perx/internal/logger"
	"perx/internal/queue"
	"perx/internal/worker"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"perx/docs"
)

//	@title			PERX Queue API
//	@version		1.0.0
//	@description	PERX Queue API

// @BasePath	/
func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	lg := logger.NewLogger(cfg.LoggerLevel)
	lg.Debug("Debug messages are available")
	lg.Info("Info messages are available")
	lg.Warn("Warn messages are available")
	lg.Error("Error messages are available")

	lg.Debug("service config", slog.Any("config", cfg))

	stopCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	taskQueue := queue.NewQueue(stopCtx, lg)
	go taskQueue.StartQueue()

	workersPool := worker.NewWorkerPool(stopCtx, lg, cfg.Workers, taskQueue)
	workersPool.StartPool()

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	lg.Info("middleware successfully connected")

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	httpHandlers := handlers.NewHandlers(lg, taskQueue)
	httpHandlers.ConnectHandlers(router)

	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	router.Mount("/swagger", httpSwagger.WrapHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
		IdleTimeout:  time.Minute,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			lg.Error("failed to listen and serve", sl.Err(err))
			os.Exit(1)
		}
	}()
	lg.Info("server started", slog.String("host", cfg.Host), slog.Int("port", cfg.Port))

	<-stopCtx.Done()
	lg.Info("Shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		lg.Error("server forced to shutdown", sl.Err(err))
	}
	workersPool.StopPool()
}
