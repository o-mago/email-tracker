package usecase

import (
	"context"
	"fmt"

	"github.com/go-api-template/app/domain/entity"
)

type GetUserInput struct {
	ID string
}

type GetUserOutput struct {
	User entity.User
}

func (u *UseCase) GetUser(ctx context.Context, input GetUserInput) (GetUserOutput, error) {
	const operation = "UseCase.GetUser"

	user, err := u.UsersRepository.GetUserByID(ctx, input.ID)
	if err != nil {
		return GetUserOutput{}, fmt.Errorf("%s -> %w", operation, err)
	}

	return GetUserOutput{
		User: user,
	}, nil
}
