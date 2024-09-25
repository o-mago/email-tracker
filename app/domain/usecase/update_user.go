package usecase

import (
	"context"
	"fmt"

	"github.com/go-api-template/app/domain/entity"
)

type UpdateUserInput struct {
	User entity.User
}

func (u *UseCase) UpdateUser(ctx context.Context, input UpdateUserInput) error {
	const operation = "UseCase.UpdateUser"

	err := u.UsersRepository.Update(ctx, input.User)
	if err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	return nil
}
