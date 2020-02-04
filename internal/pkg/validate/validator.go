package validate

import (
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
)

func Account(d *types.UpdateAccountData) []string {
	errorMessages := []string{}
	errorMessages = append(errorMessages, ValidateEntity(d.Entity)...)
	return errorMessages
}

func UpdateEntity(b *types.EntityData) []string {
	errorMessages := []string{}
	errorMessages = append(errorMessages, ValidateEntity(b)...)
	return errorMessages
}
