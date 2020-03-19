package util

import (
	"fmt"
	"strings"
)

// IsDecimalValid checks the num is positive value and with up to two decimal places.
func IsDecimalValid(num float64) bool {
	numArr := strings.Split(fmt.Sprintf("%g", num), ".")
	if len(numArr) == 1 {
		return true
	}
	if len(numArr) == 2 && len(numArr[1]) <= 2 {
		return true
	}
	return false
}
