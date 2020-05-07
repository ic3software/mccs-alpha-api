package es

import (
	"context"

	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/olivere/elastic/v7"
)

var UserAction = &userAction{}

type userAction struct {
	c     *elastic.Client
	index string
}

func (es *userAction) Register(client *elastic.Client) {
	es.c = client
	es.index = "logs"
}

// POST /transfers

func (es *userAction) Create(ua *types.UserAction) error {
	body := types.UserActionESRecord{
		UserID:    ua.UserID,
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
