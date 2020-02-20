package utils

import "github.com/ic3network/mccs-alpha-api/internal/app/types"

func CategoryToNames(categories []*types.Category) []string {
	names := make([]string, 0, len(categories))
	for _, c := range categories {
		names = append(names, c.Name)
	}
	return names
}
