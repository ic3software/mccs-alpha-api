package types

import (
	"errors"
	"time"

	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/spf13/viper"
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

	if query.QueryingEntityID == "" {
		errs = append(errs, errors.New("Please specify the querying_entity_id"))
	}

	return errs
}

type SearchTagQuery struct {
	Fragment string `json:"fragment"`
	Page     string `json:"page"`
	PageSize string `json:"pageSize"`
}

func (q *SearchTagQuery) Validate() []error {
	errs := []error{}

	_, err := util.ToInt(q.Page)
	if err != nil {
		errs = append(errs, err)
	}
	_, err = util.ToInt(q.PageSize)
	if err != nil {
		errs = append(errs, err)
	}

	return errs
}

func (q *SearchTagQuery) GetPage() int64 {
	page, _ := util.ToInt64(q.PageSize, 1)
	return page
}

func (q *SearchTagQuery) GetPageSize() int64 {
	pageSize, _ := util.ToInt64(q.PageSize, viper.GetInt64("page_size"))
	return pageSize
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
