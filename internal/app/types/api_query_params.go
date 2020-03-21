package types

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SearchEntityQuery struct {
	QueryingEntityID string
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

func (query *SearchEntityQuery) Validate() []error {
	errs := []error{}

	if query.FavoritesOnly == true && query.QueryingEntityID == "" {
		errs = append(errs, errors.New("Please specify the querying_entity_id."))
	}

	if !query.TaggedSince.IsZero() && len(query.Wants) == 0 && len(query.Offers) == 0 {
		errs = append(errs, errors.New("Please specify an offer or want tag."))
	}

	return errs
}

type SearchTagQuery struct {
	Fragment string `json:"fragment"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

func (q *SearchTagQuery) Validate() []error {
	errs := []error{}
	return errs
}

type SearchCategoryQuery struct {
	Fragment string `json:"fragment"`
	Prefix   string `json:"prefix"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

func (query *SearchCategoryQuery) Validate() []error {
	errs := []error{}
	return errs
}

type SearchTransferQuery struct {
	Page             int
	PageSize         int
	Status           string
	QueryingEntityID string

	QueryingAccountNumber string
	Offset                int
}

func (query *SearchTransferQuery) Validate() []error {
	errs := []error{}

	if query.QueryingEntityID == "" {
		errs = append(errs, errors.New("Please specify the querying_entity_id."))
	}
	if query.Status != "all" && query.Status != "initiated" && query.Status != "completed" && query.Status != "cancelled" {
		errs = append(errs, errors.New("Please specify valid status."))
	}

	return errs
}
