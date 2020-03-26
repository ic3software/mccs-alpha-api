package api

import (
	"net/http"
	"net/url"

	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/spf13/viper"
)

func NewSearchEntityQuery(q url.Values) (*types.SearchEntityQuery, error) {
	page, err := util.ToInt(q.Get("page"), 1)
	if err != nil {
		return nil, err
	}
	pageSize, err := util.ToInt(q.Get("page_size"), viper.GetInt("page_size"))
	if err != nil {
		return nil, err
	}
	return &types.SearchEntityQuery{
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

func NewSearchTagQuery(q url.Values) (*types.SearchTagQuery, error) {
	page, err := util.ToInt(q.Get("page"), 1)
	if err != nil {
		return nil, err
	}
	pageSize, err := util.ToInt(q.Get("page_size"), viper.GetInt("page_size"))
	if err != nil {
		return nil, err
	}
	return &types.SearchTagQuery{
		Fragment: q.Get("fragment"),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func NewSearchCategoryQuery(q url.Values) (*types.SearchCategoryQuery, error) {
	page, err := util.ToInt(q.Get("page"), 1)
	if err != nil {
		return nil, err
	}
	pageSize, err := util.ToInt(q.Get("page_size"), viper.GetInt("page_size"))
	if err != nil {
		return nil, err
	}
	return &types.SearchCategoryQuery{
		Fragment: q.Get("fragment"),
		Prefix:   q.Get("prefix"),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func NewSearchTransferQuery(q url.Values) (*types.SearchTransferQuery, []error) {
	page, err := util.ToInt(q.Get("page"), 1)
	if err != nil {
		return nil, []error{err}
	}
	pageSize, err := util.ToInt(q.Get("page_size"), viper.GetInt("page_size"))
	if err != nil {
		return nil, []error{err}
	}
	entity, err := logic.Entity.FindByStringID(q.Get("querying_entity_id"))
	if err != nil {
		return nil, []error{err}
	}
	query := &types.SearchTransferQuery{
		Page:                  page,
		PageSize:              pageSize,
		Status:                q.Get("status"),
		QueryingEntityID:      q.Get("querying_entity_id"),
		QueryingAccountNumber: entity.AccountNumber,
		Offset:                (page - 1) * pageSize,
	}

	return query, query.Validate()
}

func NewBalanceQuery(r *http.Request) (*types.BalanceQuery, []error) {
	query := types.BalanceQuery{
		QueryingEntityID: r.URL.Query().Get("querying_entity_id"),
	}
	return &query, query.Validate()
}
