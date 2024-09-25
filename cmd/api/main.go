package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/go-api-template/app"
	"github.com/go-api-template/app/config"
	"github.com/go-api-template/app/gateway/api"
	"github.com/go-api-template/app/gateway/postgres"
	"github.com/go-api-template/app/gateway/redis"
	"github.com/go-api-template/app/telemetry"
)

// Injected on build via ldflags.
var (
	BuildTime   = "undefined"
	BuildCommit = "undefined"
	BuildTag    = "undefined"
)

func main() {
	mainCtx := context.Background()

	// Config
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("failed to load configurations: %v", err)
	}

	// Logger
	telemetry.SetLogger(cfg, BuildTime, BuildCommit, BuildTag)

	// Open Telemetry
	otel, err := telemetry.NewOtel(mainCtx, cfg.Otel, string(cfg.Environment), BuildTag)
	if err != nil {
		log.Fatalf("failed to start otel: %v", err)
	}

	ctx := telemetry.ContextWithTracer(mainCtx, otel.Tracer)

	// Postgres
	postgresClient, err := postgres.New(ctx, cfg.Postgres)
	if err != nil {
		log.Fatalf("failed to start postgres: %v", err)
	}

	// Redis
	redisClient, err := redis.New(ctx, cfg.Redis)
	if err != nil {
		log.Fatalf("failed to start redis: %v", err)
	}

	// Application
	appl, err := app.New(cfg, postgresClient, redisClient)
	if err != nil {
		log.Fatalf("failed to start application: %v", err)
	}

	// Server
	server := &http.Server{
		Addr:         cfg.Server.APIAddress,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		Handler:      api.New(cfg, redisClient, appl.UseCase).Handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Graceful Shutdown
	stopCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	group, groupCtx := errgroup.WithContext(stopCtx)

	group.Go(func() error {
		log.Printf("starting api server")

		return server.ListenAndServe()
	})

	//nolint:contextcheck
	group.Go(func() error {
		<-groupCtx.Done()

		log.Printf("stopping api; interrupt signal received")

		timeoutCtx, cancel := context.WithTimeout(context.Background(), cfg.App.GracefulShutdownTimeout)
		defer cancel()

		var errs error

		if err := server.Shutdown(timeoutCtx); err != nil {
			errs = errors.Join(errs, fmt.Errorf("failed to stop server: %w", err))
		}

		if err := otel.Close(timeoutCtx); err != nil {
			errs = errors.Join(errs, fmt.Errorf("failed to stop otel: %w", err))
		}

		if err := redisClient.Close(); err != nil {
			errs = errors.Join(errs, fmt.Errorf("failed to stop redis: %w", err))
		}

		postgresClient.Close()

		return errs
	})

	if err := group.Wait(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("api exit reason: %v", err)
	}

	stop()
}
