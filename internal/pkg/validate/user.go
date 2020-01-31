package validate

import (
	"errors"
	"strconv"
	"strings"

	"github.com/ic3network/mccs-alpha-api/internal/pkg/util"
)

var (
	emailMaxLen = 100
)

func SignUp(email, password string) []error {
	errs := []error{}

	email = strings.ToLower(email)
	if email == "" {
		errs = append(errs, errors.New("Email is missing."))
	} else if len(email) > emailMaxLen {
		errs = append(errs, errors.New("Email address length cannot exceed "+strconv.Itoa(emailMaxLen)+" characters."))
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
