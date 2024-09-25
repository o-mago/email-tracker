package postgres

import (
	"context"
	"fmt"

	"github.com/go-api-template/app/domain/entity"
)

func (r *UsersRepository) Create(ctx context.Context, user entity.User) error {
	const (
		operation = "Repository.Users.Create"
		query     = `
			INSERT INTO users (id, name)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`
	)

	_, err := r.Client.Pool.Exec(
		ctx,
		query,
		user.ID,
		user.Name,
	)
	if err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	return nil
}
