package inertia

import "errors"

var (
	// ErrInvalidContextProps error.
	ErrInvalidContextProps = errors.New("inertia: could not convert context props to map")

	// ErrInvalidContextViewData error.
	ErrInvalidContextViewData = errors.New("inertia: could not convert context view data to map")
)
