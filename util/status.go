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
	statuses := strings.FieldsFunc(input, splitFn)

	for _, s := range statuses {
		if s != constant.Entity.Pending && s != constant.Entity.Accepted && s != constant.Entity.Rejected &&
			s != constant.Trading.Pending && s != constant.Trading.Accepted && s != constant.Trading.Rejected {
			return nil, errors.New("Please enter a valid status.")
		}
	}

	return statuses, nil
}
