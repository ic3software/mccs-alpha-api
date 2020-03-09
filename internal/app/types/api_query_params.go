package types

import (
	"errors"
	"net/url"
	"time"

	"github.com/ic3network/mccs-alpha-api/global/constant"
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

func NewSearchEntityQuery(q url.Values) (*SearchEntityQuery, error) {
	page, err := util.ToInt(q.Get("page"), 1)
	if err != nil {
		return nil, err
	}
	pageSize, err := util.ToInt(q.Get("page_size"), viper.GetInt("page_size"))
	if err != nil {
		return nil, err
	}
	return &SearchEntityQuery{
		QueryingEntityID: q.Get("querying_entity_id"),
		Page:             page,
		PageSize:         pageSize,
		EntityName:       q.Get("entity_name"),
		Category:         q.Get("category"),
		Offers:           util.ToSearchTags(q.Get("offers")),
		Wants:            util.ToSearchTags(q.Get("wants")),
		TaggedSince:      util.ParseTime(q.Get("tagged_since")),
		FavoritesOnly:    q.Get("favorites_only") == "true",
		Statuses: []string{
			constant.Entity.Accepted,
			constant.Trading.Pending,
			constant.Trading.Accepted,
			constant.Trading.Rejected,
		},
	}, nil
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
