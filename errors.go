package inertia

import "errors"

var (
	// ErrBadSsrStatusCode error.
	ErrBadSsrStatusCode = errors.New("inertia: bad ssr status code >= 400")

	// ErrInvalidContextValue error.
	ErrInvalidContextValue = errors.New("inertia: could not convert context value to expected type")
)
