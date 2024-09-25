package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/go-api-template/app/domain/erring"
)

func (c *Client) Get(ctx context.Context, key string, objByRef any) error {
	const operation = "Redis.Get"

	res, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, errCacheKeyDoesNotExist) {
			return fmt.Errorf("%s (%s) -> %w", operation, key, erring.ErrCacheKeyDoesNotExist)
		}

		return fmt.Errorf("%s (%s) -> %w", operation, key, err)
	}

	err = json.Unmarshal([]byte(res), &objByRef)
	if err != nil {
		return fmt.Errorf("%s (%s) -> %w", operation, key, err)
	}

	return nil
}

func (c *Client) Set(ctx context.Context, key string, obj any, ttl time.Duration) error {
	const operation = "Redis.Set"

	bytes, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("%s (%s) -> %w", operation, key, err)
	}

	err = c.Client.Set(ctx, key, string(bytes), ttl).Err()
	if err != nil {
		return fmt.Errorf("%s (%s) -> %w", operation, key, err)
	}

	return nil
}
