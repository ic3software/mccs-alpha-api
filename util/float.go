package util

import (
	"strconv"
)

// GET /admin/entities

func ToFloat64(input string) (*float64, error) {
	if input == "" {
		return nil, nil
	}

	s, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return nil, err
	}

	return &s, nil
}
