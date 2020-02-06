package validate

import (
	"errors"
	"strconv"
	"strings"

	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/util"
	"github.com/spf13/viper"
)

var (
	emailMaxLen = viper.GetInt("validate.email.maxLen")
)

func SignUp(req types.SignupRequest) []error {
	errs := []error{}

	req.Email = strings.ToLower(req.Email)
	if req.Email == "" {
		errs = append(errs, errors.New("Email is missing."))
	} else if len(req.Email) > emailMaxLen {
		errs = append(errs, errors.New("Email address length cannot exceed "+strconv.Itoa(emailMaxLen)+" characters."))
	} else if util.IsInValidEmail(req.Email) {
		errs = append(errs, errors.New("Email is invalid."))
	}

	if req.Password == "" {
		errs = append(errs, errors.New("Password is missing."))
	} else {
		errs = append(errs, validatePassword(req.Password)...)
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
