package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SearchEntityQuery struct {
	Page             int
	PageSize         int
	EntityName       string
	Wants            []string
	Offers           []string
	Category         string
	FavoriteEntities []primitive.ObjectID
	FavoritesOnly    bool
	TaggedSince      time.Time
	Statuses         []string // accepted", "pending", rejected", "tradingPending", "tradingAccepted", "tradingRejected"

	LocationCountry string
	LocationCity    string
}

type SearchTagQuery struct {
	Fragment string `json:"fragment"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}
