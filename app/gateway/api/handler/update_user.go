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

func (h *Handler) UpdateUserSetup(router chi.Router) {
	const (
		command = "create-user"
		pattern = "/user/{id}"
	)

	circuit := h.circuitManager.MustCreateCircuit(command)
	handler := rest.HandleWithCircuit(circuit, h.cfg.CircuitBreaker, h.cache, pattern, h.updateUser)

	router.Put(pattern, handler)
}

func (h *Handler) updateUser(req *http.Request) *response.Response {
	var request schema.UpdateUserRequest

	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		return response.AppError(errors.Join(err, erring.ErrRequestInvalid))
	}
	defer req.Body.Close()

	input := usecase.UpdateUserInput{
		User: entity.User{
			Name: request.Name,
		},
	}

	err := h.useCase.UpdateUser(req.Context(), input)
	if err != nil {
		return response.AppError(err)
	}

	return response.OK(nil)
}
