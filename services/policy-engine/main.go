package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/kenlau201/sovereign-trade-infrastructure/services/policy-engine/internal/handlers"
	"github.com/kenlau201/sovereign-trade-infrastructure/services/policy-engine/internal/policy"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Initialize Policy Engine
	engine, err := policy.NewPolicyEngine(logger)
	if err != nil {
		logger.Fatal("Failed to initialize policy engine", zap.Error(err))
	}

	// Create HTTP router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","service":"policy-engine"}`)
	})

	// Metrics
	r.Handle("/metrics", promhttp.Handler())

	// Policy evaluation endpoints
	h := handlers.NewPolicyHandler(engine, logger)
	r.Post("/evaluate", h.EvaluatePolicy)
	r.Get("/policies", h.ListPolicies)
	r.Post("/policies/reload", h.ReloadPolicies)

	// Start server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("Starting Policy Engine", zap.String("addr", server.Addr))

	// Start server in background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server error", zap.Error(err))
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Shutdown error", zap.Error(err))
	}

	logger.Info("Policy Engine stopped")
}
