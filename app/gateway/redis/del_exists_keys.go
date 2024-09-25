package redis

import (
	"context"
	"fmt"
)

func (c *Client) Del(ctx context.Context, key string) (bool, error) {
	const operation = "Redis.Del"

	count, err := c.Client.Del(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("%s (%s) -> %w", operation, key, err)
	}

	return count > 0, nil
}

func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	const operation = "Redis.Exists"

	count, err := c.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("%s (%s) -> %w", operation, key, err)
	}

	return count > 0, nil
}

func (c *Client) Keys(ctx context.Context, pattern string) ([]string, error) {
	const operation = "Redis.Keys"

	keys, err := c.Client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("%s (%s) -> %w", operation, pattern, err)
	}

	return keys, nil
}
