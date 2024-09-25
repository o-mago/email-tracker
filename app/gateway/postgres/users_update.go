package postgres

import (
	"context"
	"fmt"

	"github.com/go-api-template/app/domain/entity"
)

func (r *UsersRepository) Update(ctx context.Context, user entity.User) error {
	const (
		operation = "Repository.Users.Update"
		query     = `
			UPDATE users SET 
				name = $1,
			WHERE id = $2
		`
	)

	_, err := r.Client.Pool.Exec(
		ctx,
		query,
		user.Name,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	return nil
}
