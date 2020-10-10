package inertia

import "errors"

var (
	ErrInvalidContextProps    = errors.New("inertia: could not convert context props to map")
	ErrInvalidContextViewData = errors.New("inertia: could not convert context view data to map")
)
