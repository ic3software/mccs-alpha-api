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

	errs = append(errs, checkEmail(req.Email)...)
	errs = append(errs, validatePassword(req.Password)...)
	errs = append(errs, checkUser(types.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Telephone: req.UserPhone,
	})...)

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
	errs = append(errs, validatePassword(password)...)
	return errs
}

func UpdateUser(update types.UpdateUser) []error {
	errs := []error{}

	errs = append(errs, checkUser(types.User{
		FirstName: update.FirstName,
		LastName:  update.LastName,
		Telephone: update.UserPhone,
	})...)

	return errs
}

func checkEmail(email string) []error {
	errs := []error{}
	email = strings.ToLower(email)
	if email == "" {
		errs = append(errs, errors.New("Email is missing."))
	} else if len(email) > emailMaxLen {
		errs = append(errs, errors.New("Email address length cannot exceed "+strconv.Itoa(emailMaxLen)+" characters."))
	} else if util.IsInValidEmail(email) {
		errs = append(errs, errors.New("Email is invalid."))
	}
	return errs
}

func checkUser(user types.User) []error {
	errs := []error{}
	if len(user.FirstName) > 100 {
		errs = append(errs, errors.New("First name length cannot exceed 100 characters."))
	}
	if len(user.LastName) > 100 {
		errs = append(errs, errors.New("Last name length cannot exceed 100 characters."))
	}
	if len(user.Telephone) > 25 {
		errs = append(errs, errors.New("Telephone length cannot exceed 25 characters."))
	}
	return errs
}
