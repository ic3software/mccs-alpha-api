package validate

import (
	"strings"

	"github.com/ic3network/mccs-alpha-api/internal/pkg/util"
)

func SignUp(email, password string) []string {
	errorMessages := []string{}

	email = strings.ToLower(email)
	if email == "" {
		errorMessages = append(errorMessages, "Email is missing.")
	} else if !util.IsValidEmail(email) {
		errorMessages = append(errorMessages, "Email is invalid.")
	} else if len(email) > 100 {
		errorMessages = append(errorMessages, "Email cannot exceed 100 characters.")
	}

	return errorMessages
}
