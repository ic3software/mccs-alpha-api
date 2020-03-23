package api

import (
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NewUserRespond(user *types.User) *types.UserRespond {
	return &types.UserRespond{
		ID:                            user.ID.Hex(),
		Email:                         user.Email,
		UserPhone:                     user.Telephone,
		FirstName:                     user.FirstName,
		LastName:                      user.LastName,
		LastLoginIP:                   user.LastLoginIP,
		LastLoginDate:                 user.LastLoginDate,
		DailyEmailMatchNotification:   util.ToBool(user.DailyNotification),
		ShowTagsMatchedSinceLastLogin: util.ToBool(user.ShowRecentMatchedTags),
	}
}

func NewEntityRespondWithEmail(entity *types.Entity) *types.EntityRespond {
	return &types.EntityRespond{
		ID:                 entity.ID.Hex(),
		AccountNumber:      entity.AccountNumber,
		EntityName:         entity.EntityName,
		Email:              entity.Email,
		EntityPhone:        entity.EntityPhone,
		IncType:            entity.IncType,
		CompanyNumber:      entity.CompanyNumber,
		Website:            entity.Website,
		Turnover:           entity.Turnover,
		Description:        entity.Description,
		LocationAddress:    entity.LocationAddress,
		LocationCity:       entity.LocationCity,
		LocationRegion:     entity.LocationRegion,
		LocationPostalCode: entity.LocationPostalCode,
		LocationCountry:    entity.LocationCountry,
		Status:             entity.Status,
		Offers:             types.TagFieldToNames(entity.Offers),
		Wants:              types.TagFieldToNames(entity.Wants),
	}
}

func NewEntityRespondWithoutEmail(entity *types.Entity) *types.EntityRespond {
	return &types.EntityRespond{
		ID:                 entity.ID.Hex(),
		AccountNumber:      entity.AccountNumber,
		EntityName:         entity.EntityName,
		EntityPhone:        entity.EntityPhone,
		IncType:            entity.IncType,
		CompanyNumber:      entity.CompanyNumber,
		Website:            entity.Website,
		Turnover:           entity.Turnover,
		Description:        entity.Description,
		LocationAddress:    entity.LocationAddress,
		LocationCity:       entity.LocationCity,
		LocationRegion:     entity.LocationRegion,
		LocationPostalCode: entity.LocationPostalCode,
		LocationCountry:    entity.LocationCountry,
		Status:             entity.Status,
		Offers:             types.TagFieldToNames(entity.Offers),
		Wants:              types.TagFieldToNames(entity.Wants),
	}
}

func NewSearchEntityRespond(entity *types.Entity, queryingEntityStatus string, favoriteEntities []primitive.ObjectID) *types.SearchEntityRespond {
	email := ""
	if util.IsTradingAccepted(entity.Status) && util.IsTradingAccepted(queryingEntityStatus) {
		email = entity.Email
	}
	return &types.SearchEntityRespond{
		ID:                 entity.ID.Hex(),
		AccountNumber:      entity.AccountNumber,
		EntityName:         entity.EntityName,
		Email:              email,
		EntityPhone:        entity.EntityPhone,
		IncType:            entity.IncType,
		CompanyNumber:      entity.CompanyNumber,
		Website:            entity.Website,
		Turnover:           entity.Turnover,
		Description:        entity.Description,
		LocationAddress:    entity.LocationAddress,
		LocationCity:       entity.LocationCity,
		LocationRegion:     entity.LocationRegion,
		LocationPostalCode: entity.LocationPostalCode,
		LocationCountry:    entity.LocationCountry,
		Status:             entity.Status,
		Offers:             types.TagFieldToNames(entity.Offers),
		Wants:              types.TagFieldToNames(entity.Wants),
		IsFavorite:         util.ContainID(favoriteEntities, entity.ID),
	}
}

func NewProposeTransferRespond(journal *types.Journal) *types.ProposeTransferRespond {
	return &types.ProposeTransferRespond{
		ID:          journal.TransferID,
		From:        journal.FromAccountNumber,
		To:          journal.ToAccountNumber,
		Amount:      journal.Amount,
		Description: journal.Description,
		Status:      journal.Status,
	}
}

// Admin

func NewAdminEntityRespond(entity *types.Entity) *types.AdminEntityRespond {
	return &types.AdminEntityRespond{
		ID:                 entity.ID.Hex(),
		AccountNumber:      entity.AccountNumber,
		EntityName:         entity.EntityName,
		Email:              entity.Email,
		EntityPhone:        entity.EntityPhone,
		IncType:            entity.IncType,
		CompanyNumber:      entity.CompanyNumber,
		Website:            entity.Website,
		Turnover:           entity.Turnover,
		Description:        entity.Description,
		LocationAddress:    entity.LocationAddress,
		LocationCity:       entity.LocationCity,
		LocationRegion:     entity.LocationRegion,
		LocationPostalCode: entity.LocationPostalCode,
		LocationCountry:    entity.LocationCountry,
		Status:             entity.Status,
		Offers:             types.TagFieldToNames(entity.Offers),
		Wants:              types.TagFieldToNames(entity.Wants),
		Categories:         entity.Categories,
	}
}
