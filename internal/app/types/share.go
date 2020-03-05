package types

import (
	"errors"
	"strconv"
	"strings"

	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/spf13/viper"
)

func validateEmail(email string) []error {
	errs := []error{}
	email = strings.ToLower(email)
	emailMaxLen := viper.GetInt("validate.email.maxLen")
	if email == "" {
		errs = append(errs, errors.New("Email is missing."))
	} else if len(email) > emailMaxLen {
		errs = append(errs, errors.New("Email address length cannot exceed "+strconv.Itoa(emailMaxLen)+" characters."))
	} else if util.IsInValidEmail(email) {
		errs = append(errs, errors.New("Email is invalid."))
	}
	return errs
}
