package schema

// INPUTS.
type (
	CreateUserRequest struct {
		// Nome do usuário
		Name string `json:"name" extensions:"x-order=0"`
	}
)

// RESPONSES.
type (
	CreateUserResponse struct {
		// ID do usuário criado
		ID string `json:"id" extensions:"x-order=0"`
	}
)
