package seed

import (
	"context"

	"github.com/ic3network/mccs-alpha-api/internal/app/repository/es"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
)

var ElasticSearch = elasticSearch{}

type elasticSearch struct{}

func (_ *elasticSearch) CreateEntity(
	entity *types.Entity,
	accountNumber string,
	balanceLimit types.BalanceLimit,
) error {
	balance := 0.0
	record := types.EntityESRecord{
		ID:         entity.ID.Hex(),
		Name:       entity.EntityName,
		Email:      entity.Email,
		Offers:     entity.Offers,
		Wants:      entity.Wants,
		Status:     entity.Status,
		Categories: entity.Categories,
		// Address
		LocationCity:    entity.LocationCity,
		LocationRegion:  entity.LocationRegion,
		LocationCountry: entity.LocationCountry,
		// Account
		AccountNumber: accountNumber,
		Balance:       &balance,
		MaxPosBal:     &balanceLimit.MaxPosBal,
		MaxNegBal:     &balanceLimit.MaxNegBal,
	}

	_, err := es.Client().Index().
		Index("entities").
		Id(entity.ID.Hex()).
		BodyJson(record).
		Do(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (_ *elasticSearch) CreateUser(user *types.User) error {
	uRecord := types.UserESRecord{
		UserID:    user.ID.Hex(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}
	_, err := es.Client().Index().
		Index("users").
		Id(user.ID.Hex()).
		BodyJson(uRecord).
		Do(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (_ *elasticSearch) CreateTag(tag *types.Tag) error {
	tagRecord := types.TagESRecord{
		TagID:        tag.ID.Hex(),
		Name:         tag.Name,
		OfferAddedAt: tag.OfferAddedAt,
		WantAddedAt:  tag.WantAddedAt,
	}
	_, err := es.Client().Index().
		Index("tags").
		Id(tag.ID.Hex()).
		BodyJson(tagRecord).
		Do(context.Background())
	if err != nil {
		return err
	}

	return nil
}
