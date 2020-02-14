package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EntityESRecord is the data that will store into the elastic search.
type EntityESRecord struct {
	EntityID        string      `json:"entityID,omitempty"`
	EntityName      string      `json:"entityName,omitempty"`
	Offers          []*TagField `json:"offers,omitempty"`
	Wants           []*TagField `json:"wants,omitempty"`
	LocationCity    string      `json:"locationCity,omitempty"`
	LocationCountry string      `json:"locationCountry,omitempty"`
	Status          string      `json:"status,omitempty"`
	AdminTags       []string    `json:"adminTags,omitempty"`
}

type SearchCriteria struct {
	Page             int
	PageSize         int
	Wants            []string
	Offers           []string
	Category         string
	FavoriteEntities []primitive.ObjectID
	FavoritesOnly    bool
	TaggedSince      time.Time
	Statuses         []string // accepted", "pending", rejected", "tradingPending", "tradingAccepted", "tradingRejected"

	EntityName      string
	LocationCountry string
	LocationCity    string
}

type ESFindEntityResult struct {
	IDs             []string
	NumberOfResults int
	TotalPages      int
}
