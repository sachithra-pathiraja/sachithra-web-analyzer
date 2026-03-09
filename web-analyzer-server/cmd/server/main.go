package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"web-analyzer/internal/handler"
	"web-analyzer/internal/service"
)

func main() {

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	fetchService := service.NewFetchService(client, logger)

	analyzerHandler := handler.NewAnalyzerHandler(fetchService)

	mux := http.NewServeMux()
	mux.HandleFunc("/analyzer", analyzerHandler.Analyze)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// ---- Start Server ----
	go func() {
		logger.Info("server starting", "port", 8080)

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
