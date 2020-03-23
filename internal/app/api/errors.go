package api

import "errors"

type httpError struct {
	Message string `json:"message"`
}

type httpErrors struct {
	Errors []httpError `json:"errors"`
}

var (
	// ErrUnauthorized occurs when the user is unauthorized.
	ErrUnauthorized = errors.New("Could not authenticate you.")
	// ErrPermissionDenied occurs when the user does not have permission to perform the action.
	ErrPermissionDenied = errors.New("Permission denied.")
)
