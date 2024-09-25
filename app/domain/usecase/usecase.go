package usecase

import (
	"context"

	"github.com/go-api-template/app/domain/entity"
)

type UseCase struct {
	AppName string

	// Repos
	UsersRepository usersRepository
}

type usersRepository interface {
	Create(ctx context.Context, user entity.User) error
	GetUserByID(ctx context.Context, id string) (entity.User, error)
	Update(ctx context.Context, user entity.User) error
}
