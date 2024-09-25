package schema

// INPUTS.
type (
	UpdateUserRequest struct {
		// Nome do usuário
		Name string `json:"name" extensions:"x-order=0"`
	}
)
