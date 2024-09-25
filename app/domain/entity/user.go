package entity

import "time"

type User struct {
	ID   string
	Name string

	CreatedAt time.Time
	UpdatedAt time.Time
}
