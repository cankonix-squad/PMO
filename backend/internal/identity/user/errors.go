package user

import (
	"errors"
)

// ErrEmailTaken is returned when the requested email is already in use.
var ErrEmailTaken = errors.New("email already in use")
