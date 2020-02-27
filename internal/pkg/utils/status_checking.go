package utils

import (
	"github.com/ic3network/mccs-alpha-api/global/constant"
)

// IsAcceptedStatus checks if the entity status is accpeted.
func IsAcceptedStatus(status string) bool {
	if status == constant.Entity.Accepted ||
		status == constant.Trading.Pending ||
		status == constant.Trading.Accepted ||
		status == constant.Trading.Rejected {
		return true
	}
	return false
}
