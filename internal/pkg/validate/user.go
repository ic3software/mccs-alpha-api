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
	} else if !util.IsValidEmail(email) {
		errs = append(errs, errors.New("Email is invalid."))
	} else if len(email) > 100 {
		errs = append(errs, errors.New("Email cannot exceed 100 characters."))
	}

	return errs
}
