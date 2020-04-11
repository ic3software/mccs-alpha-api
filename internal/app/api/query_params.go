package api

import (
	"net/http"
	"net/url"

	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/spf13/viper"
)

func NewSearchTransferQuery(q url.Values) (*types.SearchTransferReqBody, []error) {
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
	query := &types.SearchTransferReqBody{
		Page:                  page,
		PageSize:              pageSize,
		Status:                q.Get("status"),
		QueryingEntityID:      q.Get("querying_entity_id"),
		QueryingAccountNumber: entity.AccountNumber,
		Offset:                (page - 1) * pageSize,
	}

	return query, query.Validate()
}

func NewBalanceQuery(r *http.Request) (*types.BalanceReqBody, []error) {
	req := types.BalanceReqBody{
		QueryingEntityID: r.URL.Query().Get("querying_entity_id"),
	}
	return &req, req.Validate()
}
