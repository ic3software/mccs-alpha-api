package types

import "time"

type SearchEntityQuery struct {
	Page          int       `json:"page"`
	PageSize      int       `json:"pageSize"`
	Category      string    `json:"category"`
	FavoritesOnly bool      `json:"favoritesOnly"`
	Offers        []string  `json:"offers"`
	Wants         []string  `json:"wants"`
	TaggedSince   time.Time `json:"taggedSince"`
}
