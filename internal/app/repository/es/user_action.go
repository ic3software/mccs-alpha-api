package es

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/olivere/elastic/v7"
)

var UserAction = &userAction{}

type userAction struct {
	c     *elastic.Client
	index string
}

func (es *userAction) Register(client *elastic.Client) {
	es.c = client
	es.index = "useractions"
}

func (es *userAction) Create(ua *types.UserAction) error {
	body := types.UserActionESRecord{
		UserID:    ua.UserID.Hex(),
		Email:     ua.Email,
		Action:    ua.Action,
		Category:  ua.Category,
		Detail:    ua.Detail,
		CreatedAt: ua.CreatedAt,
	}
	_, err := es.c.Index().
		Index(es.index).
		Id(ua.ID.Hex()).
		BodyJson(body).
		Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

// GET /admin/logs

func (es *userAction) Search(req *types.AdminSearchLogReq) (*types.ESSearchUserActionResult, error) {
	var userActions []*types.UserActionESRecord

	q := elastic.NewBoolQuery()

	if req.Email != "" {
		q.Must(elastic.NewTermQuery("email", req.Email))
	}
	if req.Action != "" {
		q.Must(newFuzzyWildcardQuery("action", req.Action))
	}
	if req.Detail != "" {
		q.Must(newFuzzyWildcardQuery("detail", req.Detail))
	}
	seachByCateogry(q, req.Categories)
	es.seachByTime(q, req.DateFrom, req.DateTo)

	res, err := es.c.Search().
		Index(es.index).
		From(req.Offset).
		Size(req.PageSize).
		Query(q).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	for _, hit := range res.Hits.Hits {
		var record types.UserActionESRecord
		err := json.Unmarshal(hit.Source, &record)
		if err != nil {
			return nil, err
		}
		userActions = append(userActions, &record)
	}

	numberOfResults := int(res.Hits.TotalHits.Value)
	totalPages := util.GetNumberOfPages(numberOfResults, req.PageSize)

	return &types.ESSearchUserActionResult{
		UserActions:     userActions,
		NumberOfResults: int(numberOfResults),
		TotalPages:      totalPages,
	}, nil
}

func seachByCateogry(q *elastic.BoolQuery, categories []string) *elastic.BoolQuery {
	if len(categories) != 0 {
		qq := elastic.NewBoolQuery()
		for _, category := range categories {
			qq.Should(elastic.NewMatchQuery("category", category))
		}
		q.Must(qq)
	}
	return q
}

func (es *userAction) seachByTime(q *elastic.BoolQuery, dateFrom time.Time, dateTo time.Time) {
	if !dateFrom.IsZero() {
		rangeQ := elastic.NewRangeQuery("createdAt").From(dateFrom)
		if !dateTo.IsZero() {
			rangeQ.To(dateTo)
		}
		qq := elastic.NewBoolQuery().Filter(rangeQ)
		q.Must(qq)
	}
}
