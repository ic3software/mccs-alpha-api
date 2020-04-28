package util

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

var emailRe = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func ValidateEmail(email string) []error {
	errs := []error{}
	email = strings.ToLower(email)
	emailMaxLen := viper.GetInt("validate.email.maxLen")
	if email == "" {
		errs = append(errs, errors.New("Email is missing."))
	} else if len(email) > emailMaxLen {
		errs = append(errs, errors.New("Email address length cannot exceed "+strconv.Itoa(emailMaxLen)+" characters."))
	} else if IsInValidEmail(email) {
		errs = append(errs, errors.New("Email is invalid."))
	}
	return errs
}

func IsValidEmail(email string) bool {
	if email == "" || !emailRe.MatchString(email) || len(email) > 100 {
		return false
	}
	return true
}

func IsInValidEmail(email string) bool {
	if email == "" || !emailRe.MatchString(email) || len(email) > 100 {
		return true
	}
	return false
}
