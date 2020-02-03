package validate

import (
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
)

func Account(d *types.UpdateAccountData) []string {
	errorMessages := []string{}
	errorMessages = append(errorMessages, ValidateBusiness(d.Business)...)
	return errorMessages
}

func UpdateBusiness(b *types.BusinessData) []string {
	errorMessages := []string{}
	errorMessages = append(errorMessages, ValidateBusiness(b)...)
	return errorMessages
}
