package middleware

import (
	"github.com/go-chi/chi/v5/middleware"
)

var (
	// chi.
	Logger       = middleware.Logger
	CleanPath    = middleware.CleanPath
	StripSlashes = middleware.StripSlashes
)
