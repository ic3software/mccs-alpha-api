package es

import (
	"context"

	"github.com/ic3network/mccs-alpha-api/internal/app/types"
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
