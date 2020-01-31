package validate

import (
	"errors"
	"strings"

	"github.com/ic3network/mccs-alpha-api/internal/pkg/util"
)

func SignUp(email, password string) []error {
	errs := []error{}

	email = strings.ToLower(email)
	if email == "" {
		errs = append(errs, errors.New("Email is missing."))
	} else if len(email) > 100 {
		errs = append(errs, errors.New("Email address length cannot exceed 100 characters."))
	} else if util.IsInValidEmail(email) {
		errs = append(errs, errors.New("Email is invalid."))
	}

	if password == "" {
		errs = append(errs, errors.New("Password is missing."))
	} else {
		errs = append(errs, validatePassword(password)...)
	}

	return errs
}

func Login(password string) []error {
	errs := []error{}

	if password == "" {
		errs = append(errs, errors.New("Password is missing."))
	}

	return errs
}

func ResetPassword(password string) []error {
	errs := []error{}

	if password == "" {
		errs = append(errs, errors.New("Password is missing."))
	} else {
		errs = append(errs, validatePassword(password)...)
	}

	return errs
}
