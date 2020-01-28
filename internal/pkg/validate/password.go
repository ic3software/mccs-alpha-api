package validate

import (
	"errors"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/passlib"
	"strconv"
	"unicode"
)

var (
	minLen     = 8
	hasLetter  = false
	hasNumber  = false
	hasSpecial = false
)

// Validate validates the given password.
func validatePassword(password string) []error {
	errs := []error{}

	for _, ch := range password {
		switch {
		case unicode.IsLetter(ch):
			hasLetter = true
		case unicode.IsNumber(ch):
			hasNumber = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	if len(password) < minLen {
		errs = append(errs, errors.New("Password must be at least "+strconv.Itoa(minLen)+" characters long."))
	} else if len(password) > 100 {
		errs = append(errs, errors.New("Password cannot exceed 100 characters."))
	}
	if !hasLetter {
		errs = append(errs, errors.New("Password must have at least one letter."))
	}
	if !hasNumber {
		errs = append(errs, errors.New("Password must have at least one numeric value."))
	}
	if !hasSpecial {
		errs = append(errs, errors.New("Password must have at least one special character."))
	}

	return errs
}

// ==================== OLD CODE =======================
func ValidatePassword(password string, confirmPassword string) []string {
	errorMessages := []string{}

	if password == "" {
		errorMessages = append(errorMessages, "Please enter a password.")
	} else if password != confirmPassword {
		errorMessages = append(errorMessages, "Password and confirmation password do not match.")
	} else {
		errorMessages = append(errorMessages, passlib.Validate(password)...)
	}

	return errorMessages
}

func validateUpdatePassword(currentPass string, newPass string, confirmPass string) []string {
	errorMessages := []string{}

	if currentPass == "" && newPass == "" && confirmPass == "" {
		return errorMessages
	}

	if currentPass == "" {
		errorMessages = append(errorMessages, "Please enter your current password.")
	} else if newPass != confirmPass {
		errorMessages = append(errorMessages, "New password and confirmation password do not match.")
	} else {
		errorMessages = append(errorMessages, passlib.Validate(newPass)...)
	}

	return errorMessages
}
