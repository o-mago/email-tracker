package postgres

type UsersRepository struct {
	*Client
}

func NewUsersRepository(client *Client) *UsersRepository {
	return &UsersRepository{client}
}
