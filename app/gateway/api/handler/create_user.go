package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/go-api-template/app/domain/entity"
	"github.com/go-api-template/app/domain/erring"
	"github.com/go-api-template/app/domain/usecase"
	"github.com/go-api-template/app/gateway/api/handler/schema"
	"github.com/go-api-template/app/gateway/api/rest"
	"github.com/go-api-template/app/gateway/api/rest/response"
)

func (h *Handler) CreateUserSetup(router chi.Router) {
	const (
		command = "create-user"
		pattern = "/user"
	)

	circuit := h.circuitManager.MustCreateCircuit(command)
	handler := rest.HandleWithCircuit(circuit, h.cfg.CircuitBreaker, h.cache, pattern, h.createUser)

	router.Post(pattern, handler)
}

func (h *Handler) createUser(req *http.Request) *response.Response {
	var request schema.CreateUserRequest

	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		return response.AppError(errors.Join(err, erring.ErrRequestInvalid))
	}
	defer req.Body.Close()

	input := usecase.CreateUserInput{
		User: entity.User{
			Name: request.Name,
		},
	}

	output, err := h.useCase.CreateUser(req.Context(), input)
	if err != nil {
		return response.AppError(err)
	}

	return response.OK(schema.CreateUserResponse{ID: output.UserID})
}
