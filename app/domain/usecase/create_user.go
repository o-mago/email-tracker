package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/go-api-template/app/domain/entity"
)

type CreateUserInput struct {
	User entity.User
}

type CreateUserOutput struct {
	UserID string
}

func (u *UseCase) CreateUser(ctx context.Context, input CreateUserInput) (CreateUserOutput, error) {
	const operation = "UseCase.CreateUser"

	uuid, _ := uuid.NewV7()

	input.User.ID = uuid.String()

	err := u.UsersRepository.Create(ctx, input.User)
	if err != nil {
		return CreateUserOutput{}, fmt.Errorf("%s -> %w", operation, err)
	}

	return CreateUserOutput{
		UserID: input.User.ID,
	}, nil
}
