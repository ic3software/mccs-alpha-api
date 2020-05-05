package util

import (
	"strconv"
)

func PointerToInt(pointer *int) int {
	if pointer != nil {
		return *pointer
	}
	return 0
}

func ToInt(input string, defaultValue ...int) (int, error) {
	if input == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return 1, nil
	}

	integer, err := strconv.Atoi(input)
	if err != nil {
		return 0, err
	}

	return integer, nil
}

func ToInt64(input string, defaultValue ...int64) (int64, error) {
	if input == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return 1, nil
	}

	integer, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return 0, err
	}

	return integer, nil
}
