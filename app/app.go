package app

import (
	"github.com/go-api-template/app/config"
	"github.com/go-api-template/app/domain/usecase"
	"github.com/go-api-template/app/gateway/postgres"
	"github.com/go-api-template/app/gateway/redis"
)

type App struct {
	UseCase *usecase.UseCase
}

func New(config config.Config, db *postgres.Client, redisClient *redis.Client) (*App, error) { //nolint: revive
	useCase := &usecase.UseCase{
		AppName:         config.App.Name,
		UsersRepository: postgres.NewUsersRepository(db),
	}

	return &App{
		UseCase: useCase,
	}, nil
}
