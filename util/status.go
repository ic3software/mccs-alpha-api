package util

import (
	"errors"
	"strings"

	"github.com/ic3network/mccs-alpha-api/global/constant"
)

func AdminMapEntityStatus(input string) ([]string, error) {
	splitFn := func(c rune) bool {
		return c == ',' || c == ' '
	}
	inputStatuses := strings.FieldsFunc(strings.ToLower(input), splitFn)

	for _, s := range inputStatuses {
		if s != "pending" && s != "accepted" && s != "rejected" {
			return nil, errors.New("Please enter a valid status.")
		}
	}

	statuses := []string{}
	for _, s := range inputStatuses {
		if s == "pending" {
			statuses = append(statuses, constant.Entity.Pending, constant.Trading.Pending)
		}
		if s == "accepted" {
			statuses = append(statuses, constant.Entity.Accepted, constant.Trading.Accepted)
		}
		if s == "rejected" {
			statuses = append(statuses, constant.Entity.Rejected, constant.Trading.Rejected)
		}
	}

	return statuses, nil
}
