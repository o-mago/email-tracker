package redis

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/pkg/errors"

	"github.com/go-api-template/app/domain/erring"
)

func (c *Client) GetResp(ctx context.Context, key string) (*http.Response, error) {
	const operation = "Redis.GetResp"

	res, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, errCacheKeyDoesNotExist) {
			return nil, fmt.Errorf("%s (%s) -> %w", operation, key, erring.ErrCacheKeyDoesNotExist)
		}

		return nil, fmt.Errorf("%s (%s) -> %w", operation, key, err)
	}

	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader([]byte(res))), nil)
	if err != nil {
		return nil, fmt.Errorf("%s (%s) -> %w", operation, key, err)
	}

	return resp, nil
}

func (c *Client) SetResp(ctx context.Context, key string, resp *http.Response, ttl time.Duration) error {
	const operation = "Redis.SetResp"

	bytes, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return fmt.Errorf("%s (%s) -> %w", operation, key, err)
	}

	err = c.Client.Set(ctx, key, string(bytes), ttl).Err()
	if err != nil {
		return fmt.Errorf("%s (%s) -> %w", operation, key, err)
	}

	return nil
}
