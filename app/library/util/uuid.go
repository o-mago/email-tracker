package util

import (
	"context"

	"github.com/google/uuid"
)

var uuidFunc = uuid.NewString

// Generates a uuid in loop until the coliderChecker returns an error, meaning no record has this uuid yet.
func GenerateUniqueUUID[T any](ctx context.Context, coliderChecker func(context.Context, string) (T, error)) string {
	for {
		uuid := uuidFunc()

		_, err := coliderChecker(ctx, uuid)
		if err != nil {
			return uuid
		}
	}
}
