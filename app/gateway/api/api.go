package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/go-api-template/app/config"
	"github.com/go-api-template/app/domain/usecase"
	"github.com/go-api-template/app/gateway/api/handler"
	"github.com/go-api-template/app/gateway/api/middleware"
	"github.com/go-api-template/app/gateway/redis"
)

type API struct {
	Handler     http.Handler
	cfg         config.Config
	useCase     *usecase.UseCase
	redisClient *redis.Client
}

func BasicHandler() http.Handler {
	router := chi.NewMux()
	handler.RegisterHealthCheckRoute(router)

	return router
}

func New(cfg config.Config, redisClient *redis.Client, useCase *usecase.UseCase) *API {
	api := &API{
		cfg:         cfg,
		useCase:     useCase,
		redisClient: redisClient,
	}

	api.setupRouter()

	return api
}

func (api *API) setupRouter() {
	router := chi.NewRouter()

	if api.cfg.Development {
		router.Use(middleware.Logger)
	}

	router.Use(
		middleware.CORS,
		middleware.CleanPath,
		middleware.StripSlashes,
		middleware.HeadersToContext,
		middleware.Recoverer,
	)

	api.registerRoutes(router)

	api.Handler = router
}

func (api *API) registerRoutes(router *chi.Mux) {
	handler.RegisterHealthCheckRoute(router)

	router.Route("/api/v1/chatbot", func(publicRouter chi.Router) {
		handler.RegisterPublicRoutes(
			publicRouter,
			api.cfg,
			api.useCase,
			api.redisClient,
		)
	})
}
