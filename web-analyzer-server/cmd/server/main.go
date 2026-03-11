package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"web-analyzer/internal/config"
	"web-analyzer/internal/handler"
	"web-analyzer/internal/middleware"
	"web-analyzer/internal/service"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg, err := config.Load("config.properties")
	if err != nil {
		logger.Error("config loading failed", "error", err)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	fetchService := service.NewFetchService(client, logger, cfg.LinkWorkers)

	analyzerHandler := handler.NewAnalyzerHandler(fetchService)

	mux := http.NewServeMux()
	mux.HandleFunc("/analyzer", analyzerHandler.Analyze)

	handlerWithMiddleware := middleware.Chain(
		mux,
		middleware.Recovery(logger),
		middleware.Logging(logger),
	)

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: handlerWithMiddleware,
	}

	// ---- Start Server ----
	go func() {
		logger.Info("server starting", "port", cfg.ServerPort)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// ---- Wait for Shutdown Signal ----
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	logger.Info("shutdown signal received")

	// ---- Graceful shutdown ----
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}

	logger.Info("server stopped gracefully")
}
