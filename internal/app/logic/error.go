package logic

import (
	"errors"
)

var (
	ErrLoginLocked = errors.New("Your account has been temporarily locked for 15 minutes. Please try again later.")
)
