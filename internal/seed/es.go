package seed

import (
	"context"

	"github.com/ic3network/mccs-alpha-api/internal/app/repository/es"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
)

var ElasticSearch = elasticSearch{}

type elasticSearch struct{}

func (_ *elasticSearch) CreateEntity(entity *types.Entity) error {
	record := types.EntityESRecord{
		EntityID:        entity.ID.Hex(),
		EntityName:      entity.EntityName,
		Offers:          entity.Offers,
		Wants:           entity.Wants,
		LocationCity:    entity.LocationCity,
		LocationCountry: entity.LocationCountry,
		Status:          entity.Status,
		Categories:      entity.Categories,
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
