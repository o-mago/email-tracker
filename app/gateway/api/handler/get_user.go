package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/go-api-template/app/domain/usecase"
	"github.com/go-api-template/app/gateway/api/rest"
	"github.com/go-api-template/app/gateway/api/rest/response"
)

func (h *Handler) GetUserSetup(router chi.Router) {
	const (
		command = "get-user"
		pattern = "/user/{id}"
	)

	circuit := h.circuitManager.MustCreateCircuit(command)
	handler := rest.HandleWithCircuit(circuit, h.cfg.CircuitBreaker, h.cache, pattern, h.getUser)

	router.Get(pattern, handler)
}

func (h *Handler) getUser(req *http.Request) *response.Response {
	id := chi.URLParam(req, "id")

	input := usecase.GetUserInput{
		ID: id,
	}

	user, err := h.useCase.GetUser(req.Context(), input)
	if err != nil {
		return response.AppError(err)
	}

	return response.OK(user)
}
