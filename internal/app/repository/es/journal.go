package es

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/olivere/elastic/v7"
)

var Journal = &journal{}

type journal struct {
	c     *elastic.Client
	index string
}

func (es *journal) Register(client *elastic.Client) {
	es.c = client
	es.index = "journals"
}

// POST /transfers
// POST /admin/transfers

func (es *journal) Create(j *types.Journal) error {
	body := types.JournalESRecord{
		TransferID:        j.TransferID,
		FromAccountNumber: j.FromAccountNumber,
		ToAccountNumber:   j.ToAccountNumber,
		Status:            j.Status,
		CreatedAt:         j.CreatedAt,
	}
	_, err := es.c.Index().
		Index(es.index).
		Id(j.TransferID).
		BodyJson(body).
		Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

// PATCH /transfers/{transferID}

func (es *journal) Update(j *types.Journal) error {
	doc := map[string]interface{}{
		"status": j.Status,
	}
	_, err := es.c.Update().
		Index(es.index).
		Id(j.TransferID).
		Doc(doc).
		Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

// GET /admin/transfers

func (es *journal) AdminSearch(req *types.AdminSearchTransferReq) (*types.ESSearchJournalResult, error) {
	var ids []string

	q := elastic.NewBoolQuery()

	es.seachByAccountNumber(q, req.AccountNumber)
	es.seachByStatus(q, req.Status)
	es.seachByTime(q, req.DateFrom, req.DateTo)

	from := req.PageSize * (req.Page - 1)
	res, err := es.c.Search().
		Index(es.index).
		From(from).
		Size(req.PageSize).
		Query(q).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	for _, hit := range res.Hits.Hits {
		var record types.JournalESRecord
		err := json.Unmarshal(hit.Source, &record)
		if err != nil {
			return nil, err
		}
		ids = append(ids, record.TransferID)
	}

	numberOfResults := int(res.Hits.TotalHits.Value)
	totalPages := util.GetNumberOfPages(numberOfResults, req.PageSize)

	return &types.ESSearchJournalResult{
		IDs:             ids,
		NumberOfResults: int(numberOfResults),
		TotalPages:      totalPages,
	}, nil
}

func (es *journal) seachByAccountNumber(q *elastic.BoolQuery, accountNumber string) {
	if accountNumber != "" {
		qq := elastic.NewBoolQuery()
		qq.Should(elastic.NewMatchQuery("fromAccountNumber", accountNumber))
		qq.Should(elastic.NewMatchQuery("toAccountNumber", accountNumber))
		q.Must(qq)
	}
}

func (es *journal) seachByStatus(q *elastic.BoolQuery, status []string) {
	if len(status) != 0 {
		qq := elastic.NewBoolQuery()
		for _, status := range status {
			qq.Should(elastic.NewMatchQuery("status", constant.MapTransferType(status)))
		}
		q.Must(qq)
	}
}

func (es *journal) seachByTime(q *elastic.BoolQuery, dateFrom time.Time, dateTo time.Time) {
	if !dateFrom.IsZero() {
		rangeQ := elastic.NewRangeQuery("createdAt").From(dateFrom)
		if !dateTo.IsZero() {
			rangeQ.To(dateTo)
		}
		qq := elastic.NewBoolQuery().Filter(rangeQ)
		q.Must(qq)
	}
}
