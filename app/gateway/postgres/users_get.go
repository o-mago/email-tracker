package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/go-api-template/app/domain/entity"
	"github.com/go-api-template/app/domain/erring"
)

func (r *UsersRepository) GetUserByID(ctx context.Context, id string) (entity.User, error) {
	const (
		operation = "Repository.Users.GetUserByID"
		query     = `
			SELECT
				name,
				created_at,
				updated_at
			FROM users
			WHERE id = $1
		`
	)

	var User entity.User

	err := r.Client.Pool.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&User.Name,
		&User.CreatedAt,
		&User.UpdatedAt,
	)
	if err != nil {
		if errors.Is(pgx.ErrNoRows, err) {
			return entity.User{}, fmt.Errorf("%s -> %w", operation, erring.ErrUserNotFound)
		}

		return entity.User{}, fmt.Errorf("%s -> %w", operation, err)
	}

	return User, nil
}
