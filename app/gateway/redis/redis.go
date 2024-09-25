package redis

import (
	"context"
	"crypto/tls"
	"fmt"

	goredisotel "github.com/redis/go-redis/extra/redisotel/v9"
	goredis "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"

	"github.com/go-api-template/app/config"
)

const errCacheKeyDoesNotExist = goredis.Nil

type Client struct {
	Client *goredis.Client
}

func (c *Client) Close() error {
	return c.Client.Close() //nolint:wrapcheck
}

func New(ctx context.Context, cfg config.Redis) (*Client, error) {
	const operation = "Redis.New"

	opts := &goredis.Options{
		Addr:     cfg.Address(),
		Username: cfg.User,
		Password: cfg.Password,
	}

	if cfg.UseTLS {
		opts.TLSConfig = &tls.Config{
			ServerName: cfg.Host,
			MinVersion: tls.VersionTLS12,
		}
	}

	client := goredis.NewClient(opts)

	attrs := goredisotel.WithAttributes(
		semconv.NetPeerNameKey.String(cfg.Host),
		semconv.NetPeerPortKey.String(cfg.Port),
	)

	err := goredisotel.InstrumentTracing(
		client,
		attrs,
		goredisotel.WithTracerProvider(otel.GetTracerProvider()),
	)
	if err != nil {
		return nil, fmt.Errorf("%s -> %w", operation, err)
	}

	err = goredisotel.InstrumentMetrics(
		client,
		attrs,
		goredisotel.WithMeterProvider(otel.GetMeterProvider()),
	)
	if err != nil {
		return nil, fmt.Errorf("%s -> %w", operation, err)
	}

	// Ping using the dial connect timeout. TODO: healthcheck.
	res := client.Ping(ctx)
	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("%s -> %w", operation, err)
	}

	return &Client{
		Client: client,
	}, nil
}
