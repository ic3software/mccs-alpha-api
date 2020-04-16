package types

import (
	"time"

	"github.com/ic3network/mccs-alpha-api/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NewUserRespond(user *User) *UserRespond {
	return &UserRespond{
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

type UserRespond struct {
	ID                            string    `json:"id"`
	Email                         string    `json:"email"`
	FirstName                     string    `json:"firstName"`
	LastName                      string    `json:"lastName"`
	UserPhone                     string    `json:"userPhone"`
	LastLoginIP                   string    `json:"lastLoginIP"`
	LastLoginDate                 time.Time `json:"lastLoginDate"`
	DailyEmailMatchNotification   bool      `json:"dailyEmailMatchNotification"`
	ShowTagsMatchedSinceLastLogin bool      `json:"showTagsMatchedSinceLastLogin"`
}

func NewEntityRespondWithEmail(entity *Entity) *EntityRespond {
	return &EntityRespond{
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
		Offers:             TagFieldToNames(entity.Offers),
		Wants:              TagFieldToNames(entity.Wants),
	}
}

func NewEntityRespondWithoutEmail(entity *Entity) *EntityRespond {
	return &EntityRespond{
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
		Offers:             TagFieldToNames(entity.Offers),
		Wants:              TagFieldToNames(entity.Wants),
	}
}

type EntityRespond struct {
	ID                 string   `json:"id"`
	AccountNumber      string   `json:"accountNumber"`
	EntityName         string   `json:"entityName"`
	Email              string   `json:"email,omitempty"`
	EntityPhone        string   `json:"entityPhone"`
	IncType            string   `json:"incType"`
	CompanyNumber      string   `json:"companyNumber"`
	Website            string   `json:"website"`
	Turnover           int      `json:"turnover"`
	Description        string   `json:"description"`
	LocationAddress    string   `json:"locationAddress"`
	LocationCity       string   `json:"locationCity"`
	LocationRegion     string   `json:"locationRegion"`
	LocationPostalCode string   `json:"locationPostalCode"`
	LocationCountry    string   `json:"locationCountry"`
	Status             string   `json:"status"`
	Offers             []string `json:"offers"`
	Wants              []string `json:"wants"`
}

func NewSearchEntityRespond(entity *Entity, queryingEntityStatus string, favoriteEntities []primitive.ObjectID) *SearchEntityRespond {
	email := ""
	if util.IsTradingAccepted(entity.Status) && util.IsTradingAccepted(queryingEntityStatus) {
		email = entity.Email
	}
	return &SearchEntityRespond{
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
		Offers:             TagFieldToNames(entity.Offers),
		Wants:              TagFieldToNames(entity.Wants),
		IsFavorite:         util.ContainID(favoriteEntities, entity.ID.Hex()),
	}
}

// SearchEntityRespond will always return IsFavorite.
type SearchEntityRespond struct {
	ID                 string   `json:"id"`
	AccountNumber      string   `json:"accountNumber"`
	EntityName         string   `json:"entityName"`
	Email              string   `json:"email,omitempty"`
	EntityPhone        string   `json:"entityPhone"`
	IncType            string   `json:"incType"`
	CompanyNumber      string   `json:"companyNumber"`
	Website            string   `json:"website"`
	Turnover           int      `json:"turnover"`
	Description        string   `json:"description"`
	LocationAddress    string   `json:"locationAddress"`
	LocationCity       string   `json:"locationCity"`
	LocationRegion     string   `json:"locationRegion"`
	LocationPostalCode string   `json:"locationPostalCode"`
	LocationCountry    string   `json:"locationCountry"`
	Status             string   `json:"status"`
	Offers             []string `json:"offers"`
	Wants              []string `json:"wants"`
	IsFavorite         bool     `json:"isFavorite"`
}

func NewProposeTransferRespond(journal *Journal) *ProposeTransferRespond {
	return &ProposeTransferRespond{
		ID:          journal.TransferID,
		From:        journal.FromAccountNumber,
		To:          journal.ToAccountNumber,
		Amount:      journal.Amount,
		Description: journal.Description,
		Status:      journal.Status,
	}
}

type ProposeTransferRespond struct {
	ID          string  `json:"id"`
	From        string  `json:"from"`
	To          string  `json:"to"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
}

type Transfer struct {
	TransferID    string     `json:"id"`
	Transfer      string     `json:"transfer"`
	IsInitiator   bool       `json:"isInitiator"`
	AccountNumber string     `json:"accountNumber"`
	EntityName    string     `json:"entityName"`
	Amount        float64    `json:"amount"`
	Description   string     `json:"description"`
	Status        string     `json:"status"`
	CreatedAt     *time.Time `json:"dateProposed,omitempty"`
	CompletedAt   *time.Time `json:"dateCompleted,omitempty"`
}

type SearchTransferRespond struct {
	Transfers       []*Transfer
	NumberOfResults int
	TotalPages      int
}

// Admin

func NewAdminEntityRespond(entity *Entity) *AdminEntityRespond {
	return &AdminEntityRespond{
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
		Offers:             TagFieldToNames(entity.Offers),
		Wants:              TagFieldToNames(entity.Wants),
		Categories:         entity.Categories,
	}
}

type AdminEntityRespond struct {
	ID                 string   `json:"id"`
	AccountNumber      string   `json:"accountNumber"`
	EntityName         string   `json:"entityName"`
	Email              string   `json:"email,omitempty"`
	EntityPhone        string   `json:"entityPhone"`
	IncType            string   `json:"incType"`
	CompanyNumber      string   `json:"companyNumber"`
	Website            string   `json:"website"`
	Turnover           int      `json:"turnover"`
	Description        string   `json:"description"`
	LocationAddress    string   `json:"locationAddress"`
	LocationCity       string   `json:"locationCity"`
	LocationRegion     string   `json:"locationRegion"`
	LocationPostalCode string   `json:"locationPostalCode"`
	LocationCountry    string   `json:"locationCountry"`
	Status             string   `json:"status"`
	Offers             []string `json:"offers,omitempty"`
	Wants              []string `json:"wants,omitempty"`
	Categories         []string `json:"categories,omitempty"`
}

func NewAdminUserRespond(user *User) *AdminUserRespond {
	return &AdminUserRespond{
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

type AdminUserRespond struct {
	ID                            string    `json:"id"`
	Email                         string    `json:"email"`
	FirstName                     string    `json:"firstName"`
	LastName                      string    `json:"lastName"`
	UserPhone                     string    `json:"userPhone"`
	LastLoginIP                   string    `json:"lastLoginIP"`
	LastLoginDate                 time.Time `json:"lastLoginDate"`
	DailyEmailMatchNotification   bool      `json:"dailyEmailMatchNotification"`
	ShowTagsMatchedSinceLastLogin bool      `json:"showTagsMatchedSinceLastLogin"`
}

func NewAdminCategoryRespond(category *Category) *AdminCategoryRespond {
	return &AdminCategoryRespond{
		ID:   category.ID.Hex(),
		Name: category.Name,
	}
}

type AdminCategoryRespond struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TagRespond struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewAdminGetUserRespond(user *User, entities []*Entity) *AdminGetUserRespond {
	adminEntityResponds := []*AdminEntityRespond{}
	for _, e := range entities {
		adminEntityResponds = append(adminEntityResponds, NewAdminEntityRespond(e))
	}

	return &AdminGetUserRespond{
		ID:                            user.ID.Hex(),
		Email:                         user.Email,
		UserPhone:                     user.Telephone,
		FirstName:                     user.FirstName,
		LastName:                      user.LastName,
		LastLoginIP:                   user.LastLoginIP,
		LastLoginDate:                 user.LastLoginDate,
		DailyEmailMatchNotification:   util.ToBool(user.DailyNotification),
		ShowTagsMatchedSinceLastLogin: util.ToBool(user.ShowRecentMatchedTags),
		Entities:                      adminEntityResponds,
	}
}

type AdminGetUserRespond struct {
	ID                            string                `json:"id"`
	Email                         string                `json:"email"`
	FirstName                     string                `json:"firstName"`
	LastName                      string                `json:"lastName"`
	UserPhone                     string                `json:"userPhone"`
	LastLoginIP                   string                `json:"lastLoginIP"`
	LastLoginDate                 time.Time             `json:"lastLoginDate"`
	DailyEmailMatchNotification   bool                  `json:"dailyEmailMatchNotification"`
	ShowTagsMatchedSinceLastLogin bool                  `json:"showTagsMatchedSinceLastLogin"`
	Entities                      []*AdminEntityRespond `json:"entities"`
}

func NewAdminGetEntityRespond(
	entity *Entity,
	users []*User,
	account *Account,
	balanceLimit *BalanceLimit,
) *AdminGetEntityRespond {
	adminUserResponds := []*AdminUserRespond{}
	for _, u := range users {
		adminUserResponds = append(adminUserResponds, NewAdminUserRespond(u))
	}

	return &AdminGetEntityRespond{
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
		Offers:             TagFieldToNames(entity.Offers),
		Wants:              TagFieldToNames(entity.Wants),
		Categories:         entity.Categories,
		Users:              adminUserResponds,
		Balance:            account.Balance,
		MaxNegativeBalance: balanceLimit.MaxNegBal,
		MaxPositiveBalance: balanceLimit.MaxPosBal,
	}
}

type AdminGetEntityRespond struct {
	ID                 string              `json:"id"`
	AccountNumber      string              `json:"accountNumber"`
	EntityName         string              `json:"entityName"`
	Email              string              `json:"email,omitempty"`
	EntityPhone        string              `json:"entityPhone"`
	IncType            string              `json:"incType"`
	CompanyNumber      string              `json:"companyNumber"`
	Website            string              `json:"website"`
	Turnover           int                 `json:"turnover"`
	Description        string              `json:"description"`
	LocationAddress    string              `json:"locationAddress"`
	LocationCity       string              `json:"locationCity"`
	LocationRegion     string              `json:"locationRegion"`
	LocationPostalCode string              `json:"locationPostalCode"`
	LocationCountry    string              `json:"locationCountry"`
	Status             string              `json:"status"`
	Offers             []string            `json:"offers,omitempty"`
	Wants              []string            `json:"wants,omitempty"`
	Categories         []string            `json:"categories,omitempty"`
	Users              []*AdminUserRespond `json:"users"`
	Balance            float64             `json:"balance"`
	MaxPositiveBalance float64             `json:"maxPositiveBalance"`
	MaxNegativeBalance float64             `json:"maxNegativeBalance"`
}
