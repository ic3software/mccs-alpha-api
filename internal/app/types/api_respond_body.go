package types

import (
	"time"

	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GET /user

func NewUserRespond(user *User) *UserRespond {
	return &UserRespond{
		ID:            user.ID.Hex(),
		Email:         user.Email,
		Telephone:     user.Telephone,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		LastLoginIP:   user.LastLoginIP,
		LastLoginDate: user.LastLoginDate,
	}
}

type UserRespond struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	FirstName     string    `json:"firstName"`
	LastName      string    `json:"lastName"`
	Telephone     string    `json:"telephone"`
	LastLoginIP   string    `json:"lastLoginIP"`
	LastLoginDate time.Time `json:"lastLoginDate"`
}

// GET /user/entities

func NewEntityRespondWithEmail(entity *Entity) *EntityRespond {
	return &EntityRespond{
		ID:                                 entity.ID.Hex(),
		AccountNumber:                      entity.AccountNumber,
		Name:                               entity.Name,
		Email:                              entity.Email,
		Telephone:                          entity.Telephone,
		IncType:                            entity.IncType,
		CompanyNumber:                      entity.CompanyNumber,
		Website:                            entity.Website,
		DeclaredTurnover:                   entity.DeclaredTurnover,
		Description:                        entity.Description,
		Address:                            entity.Address,
		City:                               entity.City,
		Region:                             entity.Region,
		PostalCode:                         entity.PostalCode,
		Country:                            entity.Country,
		Status:                             entity.Status,
		ShowTagsMatchedSinceLastLogin:      util.ToBool(entity.ShowTagsMatchedSinceLastLogin),
		ReceiveDailyMatchNotificationEmail: util.ToBool(entity.ReceiveDailyMatchNotificationEmail),
		Offers:                             TagFieldToNames(entity.Offers),
		Wants:                              TagFieldToNames(entity.Wants),
		Categories:                         entity.Categories,
	}
}

func NewEntityRespondWithoutEmail(entity *Entity) *EntityRespond {
	return &EntityRespond{
		ID:                                 entity.ID.Hex(),
		AccountNumber:                      entity.AccountNumber,
		Name:                               entity.Name,
		Telephone:                          entity.Telephone,
		IncType:                            entity.IncType,
		CompanyNumber:                      entity.CompanyNumber,
		Website:                            entity.Website,
		DeclaredTurnover:                   entity.DeclaredTurnover,
		Description:                        entity.Description,
		Address:                            entity.Address,
		City:                               entity.City,
		Region:                             entity.Region,
		PostalCode:                         entity.PostalCode,
		Country:                            entity.Country,
		Status:                             entity.Status,
		ShowTagsMatchedSinceLastLogin:      util.ToBool(entity.ShowTagsMatchedSinceLastLogin),
		ReceiveDailyMatchNotificationEmail: util.ToBool(entity.ReceiveDailyMatchNotificationEmail),
		Offers:                             TagFieldToNames(entity.Offers),
		Wants:                              TagFieldToNames(entity.Wants),
		Categories:                         entity.Categories,
	}
}

type EntityRespond struct {
	ID                                 string   `json:"id"`
	AccountNumber                      string   `json:"accountNumber"`
	Name                               string   `json:"name"`
	Email                              string   `json:"email,omitempty"`
	Telephone                          string   `json:"telephone"`
	IncType                            string   `json:"incType"`
	CompanyNumber                      string   `json:"companyNumber"`
	Website                            string   `json:"website"`
	DeclaredTurnover                   *int     `json:"declaredTurnover"`
	Description                        string   `json:"description"`
	Address                            string   `json:"address"`
	City                               string   `json:"city"`
	Region                             string   `json:"region"`
	PostalCode                         string   `json:"postalCode"`
	Country                            string   `json:"country"`
	Status                             string   `json:"status"`
	ShowTagsMatchedSinceLastLogin      bool     `json:"showTagsMatchedSinceLastLogin"`
	ReceiveDailyMatchNotificationEmail bool     `json:"receiveDailyMatchNotificationEmail"`
	Offers                             []string `json:"offers"`
	Wants                              []string `json:"wants"`
	Categories                         []string `json:"categories"`
}

// GET /entities

func NewSearchEntityRespond(entity *Entity, queryingEntityStatus string, favoriteEntities []primitive.ObjectID) *SearchEntityRespond {
	email := ""
	if util.IsTradingAccepted(entity.Status) && util.IsTradingAccepted(queryingEntityStatus) {
		email = entity.Email
	}
	return &SearchEntityRespond{
		ID:               entity.ID.Hex(),
		AccountNumber:    entity.AccountNumber,
		Name:             entity.Name,
		Email:            email,
		Telephone:        entity.Telephone,
		IncType:          entity.IncType,
		CompanyNumber:    entity.CompanyNumber,
		Website:          entity.Website,
		DeclaredTurnover: entity.DeclaredTurnover,
		Description:      entity.Description,
		Address:          entity.Address,
		City:             entity.City,
		Region:           entity.Region,
		PostalCode:       entity.PostalCode,
		Country:          entity.Country,
		Status:           entity.Status,
		Offers:           TagFieldToNames(entity.Offers),
		Wants:            TagFieldToNames(entity.Wants),
		Categories:       entity.Categories,
		IsFavorite:       util.ContainID(favoriteEntities, entity.ID.Hex()),
	}
}

type SearchEntityRespond struct {
	ID               string   `json:"id"`
	AccountNumber    string   `json:"accountNumber"`
	Name             string   `json:"name"`
	Email            string   `json:"email,omitempty"`
	Telephone        string   `json:"telephone"`
	IncType          string   `json:"incType"`
	CompanyNumber    string   `json:"companyNumber"`
	Website          string   `json:"website"`
	DeclaredTurnover *int     `json:"declaredTurnover"`
	Description      string   `json:"description"`
	Address          string   `json:"address"`
	City             string   `json:"city"`
	Region           string   `json:"region"`
	PostalCode       string   `json:"postalCode"`
	Country          string   `json:"country"`
	Status           string   `json:"status"`
	Offers           []string `json:"offers"`
	Wants            []string `json:"wants"`
	Categories       []string `json:"categories"`
	IsFavorite       bool     `json:"isFavorite"`
}

// POST /transfers

func NewProposeTransferRespond(journal *Journal) *ProposeTransferRespond {
	return &ProposeTransferRespond{
		ID:          journal.TransferID,
		From:        journal.FromAccountNumber,
		To:          journal.ToAccountNumber,
		Amount:      journal.Amount,
		Description: journal.Description,
		Status:      journal.Status,
		CreatedAt:   &journal.CreatedAt,
	}
}

type ProposeTransferRespond struct {
	ID          string     `json:"id"`
	From        string     `json:"from"`
	To          string     `json:"to"`
	Amount      float64    `json:"amount"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	CreatedAt   *time.Time `json:"dateProposed,omitempty"`
}

// GET /transfers

func NewJournalsToTransfersRespond(journals []*Journal, queryingAccountNumber string) []*TransferRespond {
	transfers := []*TransferRespond{}

	for _, j := range journals {
		t := &TransferRespond{
			TransferID:  j.TransferID,
			Description: j.Description,
			Amount:      j.Amount,
			CreatedAt:   &j.CreatedAt,
			Status:      j.Status,
		}
		if j.InitiatedBy == queryingAccountNumber {
			t.IsInitiator = true
		}
		if j.FromAccountNumber == queryingAccountNumber {
			t.Transfer = "out"
			t.AccountNumber = j.ToAccountNumber
			t.EntityName = j.ToEntityName
		} else {
			t.Transfer = "in"
			t.AccountNumber = j.FromAccountNumber
			t.EntityName = j.FromEntityName
		}
		if j.Status == constant.Transfer.Completed {
			t.CompletedAt = &j.UpdatedAt
		}

		transfers = append(transfers, t)
	}

	return transfers
}

type TransferRespond struct {
	TransferID         string     `json:"id"`
	Transfer           string     `json:"transfer"`
	IsInitiator        bool       `json:"isInitiator"`
	AccountNumber      string     `json:"accountNumber"`
	EntityName         string     `json:"entityName"`
	Amount             float64    `json:"amount"`
	Description        string     `json:"description"`
	Status             string     `json:"status"`
	CancellationReason string     `json:"cancellationReason,omitempty"`
	CreatedAt          *time.Time `json:"dateProposed,omitempty"`
	CompletedAt        *time.Time `json:"dateCompleted,omitempty"`
}

type SearchTransferRespond struct {
	Transfers       []*TransferRespond
	NumberOfResults int
	TotalPages      int
}

func NewAdminEntityRespond(entity *Entity) *AdminEntityRespond {
	return &AdminEntityRespond{
		ID:                                 entity.ID.Hex(),
		AccountNumber:                      entity.AccountNumber,
		Name:                               entity.Name,
		Email:                              entity.Email,
		Telephone:                          entity.Telephone,
		IncType:                            entity.IncType,
		CompanyNumber:                      entity.CompanyNumber,
		Website:                            entity.Website,
		DeclaredTurnover:                   entity.DeclaredTurnover,
		Description:                        entity.Description,
		Address:                            entity.Address,
		City:                               entity.City,
		Region:                             entity.Region,
		PostalCode:                         entity.PostalCode,
		Country:                            entity.Country,
		Status:                             entity.Status,
		Offers:                             TagFieldToNames(entity.Offers),
		Wants:                              TagFieldToNames(entity.Wants),
		Categories:                         entity.Categories,
		ShowTagsMatchedSinceLastLogin:      util.ToBool(entity.ShowTagsMatchedSinceLastLogin),
		ReceiveDailyMatchNotificationEmail: util.ToBool(entity.ReceiveDailyMatchNotificationEmail),
	}
}

type AdminEntityRespond struct {
	ID                                 string   `json:"id"`
	AccountNumber                      string   `json:"accountNumber"`
	Name                               string   `json:"name"`
	Email                              string   `json:"email,omitempty"`
	Telephone                          string   `json:"telephone"`
	IncType                            string   `json:"incType"`
	CompanyNumber                      string   `json:"companyNumber"`
	Website                            string   `json:"website"`
	DeclaredTurnover                   *int     `json:"declaredTurnover"`
	Description                        string   `json:"description"`
	Address                            string   `json:"address"`
	City                               string   `json:"city"`
	Region                             string   `json:"region"`
	PostalCode                         string   `json:"postalCode"`
	Country                            string   `json:"country"`
	Status                             string   `json:"status"`
	Offers                             []string `json:"offers,omitempty"`
	Wants                              []string `json:"wants,omitempty"`
	Categories                         []string `json:"categories,omitempty"`
	ReceiveDailyMatchNotificationEmail bool     `json:"receiveDailyMatchNotificationEmail"`
	ShowTagsMatchedSinceLastLogin      bool     `json:"showTagsMatchedSinceLastLogin"`
}

func NewAdminUserRespond(user *User) *AdminUserRespond {
	return &AdminUserRespond{
		ID:            user.ID.Hex(),
		Email:         user.Email,
		Telephone:     user.Telephone,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		LastLoginIP:   user.LastLoginIP,
		LastLoginDate: user.LastLoginDate,
	}
}

type AdminUserRespond struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	FirstName     string    `json:"firstName"`
	LastName      string    `json:"lastName"`
	Telephone     string    `json:"telephone"`
	LastLoginIP   string    `json:"lastLoginIP"`
	LastLoginDate time.Time `json:"lastLoginDate"`
}

// Category

type CategoryRespond struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GET /categories

func NewSearchCategoryRespond(categories []*Category) []*CategoryRespond {
	result := []*CategoryRespond{}
	for _, category := range categories {
		result = append(result, NewCategoryRespond(category))
	}
	return result
}

// POST /admin/categories

func NewCategoryRespond(category *Category) *CategoryRespond {
	return &CategoryRespond{
		ID:   category.ID.Hex(),
		Name: category.Name,
	}
}

// Tag

type TagRespond struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GET /tags
// GET /admin/tags

func NewSearchTagRespond(tags []*Tag) []*TagRespond {
	result := []*TagRespond{}
	for _, tag := range tags {
		result = append(result, NewTagRespond(tag))
	}
	return result
}

// POST /admin/tags

func NewTagRespond(tag *Tag) *TagRespond {
	return &TagRespond{
		ID:   tag.ID.Hex(),
		Name: tag.Name,
	}
}

func NewAdminGetUserRespond(user *User, entities []*Entity) *AdminGetUserRespond {
	adminEntityResponds := []*AdminEntityRespond{}
	for _, e := range entities {
		adminEntityResponds = append(adminEntityResponds, NewAdminEntityRespond(e))
	}

	return &AdminGetUserRespond{
		ID:            user.ID.Hex(),
		Email:         user.Email,
		Telephone:     user.Telephone,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		LastLoginIP:   user.LastLoginIP,
		LastLoginDate: user.LastLoginDate,
		Entities:      adminEntityResponds,
	}
}

type AdminGetUserRespond struct {
	ID            string                `json:"id"`
	Email         string                `json:"email"`
	FirstName     string                `json:"firstName"`
	LastName      string                `json:"lastName"`
	Telephone     string                `json:"telephone"`
	LastLoginIP   string                `json:"lastLoginIP"`
	LastLoginDate time.Time             `json:"lastLoginDate"`
	Entities      []*AdminEntityRespond `json:"entities"`
}

// DELETE /admin/users/{userID}

func NewAdminDeleteUserRespond(user *User) *AdminDeleteUserRespond {
	return &AdminDeleteUserRespond{
		ID:            user.ID.Hex(),
		Email:         user.Email,
		Telephone:     user.Telephone,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		LastLoginIP:   user.LastLoginIP,
		LastLoginDate: user.LastLoginDate,
	}
}

type AdminDeleteUserRespond struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	FirstName     string    `json:"firstName"`
	LastName      string    `json:"lastName"`
	Telephone     string    `json:"telephone"`
	LastLoginIP   string    `json:"lastLoginIP"`
	LastLoginDate time.Time `json:"lastLoginDate"`
}

// GET /admin/entities

func NewAdminSearchEntityRespond(
	entity *Entity,
	users []*User,
	account *Account,
	balanceLimit *BalanceLimit,
) *AdminSearchEntityRespond {
	adminUserResponds := []*AdminUserRespond{}
	for _, u := range users {
		adminUserResponds = append(adminUserResponds, NewAdminUserRespond(u))
	}

	return &AdminSearchEntityRespond{
		ID:                                 entity.ID.Hex(),
		AccountNumber:                      entity.AccountNumber,
		Name:                               entity.Name,
		Email:                              entity.Email,
		Telephone:                          entity.Telephone,
		IncType:                            entity.IncType,
		CompanyNumber:                      entity.CompanyNumber,
		Website:                            entity.Website,
		DeclaredTurnover:                   entity.DeclaredTurnover,
		Description:                        entity.Description,
		Address:                            entity.Address,
		City:                               entity.City,
		Region:                             entity.Region,
		PostalCode:                         entity.PostalCode,
		Country:                            entity.Country,
		Status:                             entity.Status,
		Offers:                             TagFieldToNames(entity.Offers),
		Wants:                              TagFieldToNames(entity.Wants),
		Categories:                         entity.Categories,
		ShowTagsMatchedSinceLastLogin:      util.ToBool(entity.ShowTagsMatchedSinceLastLogin),
		ReceiveDailyMatchNotificationEmail: util.ToBool(entity.ReceiveDailyMatchNotificationEmail),
		Balance:                            account.Balance,
		MaxNegativeBalance:                 balanceLimit.MaxNegBal,
		MaxPositiveBalance:                 balanceLimit.MaxPosBal,
		Users:                              adminUserResponds,
	}
}

type AdminSearchEntityRespond struct {
	ID                                 string              `json:"id"`
	AccountNumber                      string              `json:"accountNumber"`
	Name                               string              `json:"name"`
	Email                              string              `json:"email,omitempty"`
	Telephone                          string              `json:"telephone"`
	IncType                            string              `json:"incType"`
	CompanyNumber                      string              `json:"companyNumber"`
	Website                            string              `json:"website"`
	DeclaredTurnover                   *int                `json:"declaredTurnover"`
	Description                        string              `json:"description"`
	Address                            string              `json:"address"`
	City                               string              `json:"city"`
	Region                             string              `json:"region"`
	PostalCode                         string              `json:"postalCode"`
	Country                            string              `json:"country"`
	Status                             string              `json:"status"`
	Offers                             []string            `json:"offers,omitempty"`
	Wants                              []string            `json:"wants,omitempty"`
	Categories                         []string            `json:"categories,omitempty"`
	ShowTagsMatchedSinceLastLogin      bool                `json:"showTagsMatchedSinceLastLogin"`
	ReceiveDailyMatchNotificationEmail bool                `json:"receiveDailyMatchNotificationEmail"`
	Balance                            float64             `json:"balance"`
	MaxPositiveBalance                 float64             `json:"maxPositiveBalance"`
	MaxNegativeBalance                 float64             `json:"maxNegativeBalance"`
	Users                              []*AdminUserRespond `json:"users"`
}

// GET /admin/entities/{entityID}

func NewAdminGetEntityRespond(
	entity *Entity,
	users []*User,
	account *Account,
	balanceLimit *BalanceLimit,
	pendingTransfers []*AdminTransferRespond,
) *AdminGetEntityRespond {
	adminUserResponds := []*AdminUserRespond{}
	for _, u := range users {
		adminUserResponds = append(adminUserResponds, NewAdminUserRespond(u))
	}

	return &AdminGetEntityRespond{
		ID:                                 entity.ID.Hex(),
		AccountNumber:                      entity.AccountNumber,
		Name:                               entity.Name,
		Email:                              entity.Email,
		Telephone:                          entity.Telephone,
		IncType:                            entity.IncType,
		CompanyNumber:                      entity.CompanyNumber,
		Website:                            entity.Website,
		DeclaredTurnover:                   entity.DeclaredTurnover,
		Description:                        entity.Description,
		Address:                            entity.Address,
		City:                               entity.City,
		Region:                             entity.Region,
		PostalCode:                         entity.PostalCode,
		Country:                            entity.Country,
		Status:                             entity.Status,
		Offers:                             TagFieldToNames(entity.Offers),
		Wants:                              TagFieldToNames(entity.Wants),
		Categories:                         entity.Categories,
		ShowTagsMatchedSinceLastLogin:      util.ToBool(entity.ShowTagsMatchedSinceLastLogin),
		ReceiveDailyMatchNotificationEmail: util.ToBool(entity.ReceiveDailyMatchNotificationEmail),
		Balance:                            account.Balance,
		MaxNegativeBalance:                 balanceLimit.MaxNegBal,
		MaxPositiveBalance:                 balanceLimit.MaxPosBal,
		PendingTransfers:                   pendingTransfers,
		Users:                              adminUserResponds,
	}
}

type AdminGetEntityRespond struct {
	ID                                 string                  `json:"id"`
	AccountNumber                      string                  `json:"accountNumber"`
	Name                               string                  `json:"name"`
	Email                              string                  `json:"email,omitempty"`
	Telephone                          string                  `json:"telephone"`
	IncType                            string                  `json:"incType"`
	CompanyNumber                      string                  `json:"companyNumber"`
	Website                            string                  `json:"website"`
	DeclaredTurnover                   *int                    `json:"declaredTurnover"`
	Description                        string                  `json:"description"`
	Address                            string                  `json:"address"`
	City                               string                  `json:"city"`
	Region                             string                  `json:"region"`
	PostalCode                         string                  `json:"postalCode"`
	Country                            string                  `json:"country"`
	Status                             string                  `json:"status"`
	Offers                             []string                `json:"offers,omitempty"`
	Wants                              []string                `json:"wants,omitempty"`
	Categories                         []string                `json:"categories,omitempty"`
	ShowTagsMatchedSinceLastLogin      bool                    `json:"showTagsMatchedSinceLastLogin"`
	ReceiveDailyMatchNotificationEmail bool                    `json:"receiveDailyMatchNotificationEmail"`
	Balance                            float64                 `json:"balance"`
	MaxPositiveBalance                 float64                 `json:"maxPositiveBalance"`
	MaxNegativeBalance                 float64                 `json:"maxNegativeBalance"`
	PendingTransfers                   []*AdminTransferRespond `json:"pendingTransfers"`
	Users                              []*AdminUserRespond     `json:"users"`
}

// PATCH /admin/entities/{entityID}

func NewAdminUpdateEntityRespond(users []*User, entity *Entity, balanceLimit *BalanceLimit) *AdminUpdateEntityRespond {
	adminUserResponds := []*AdminUserRespond{}
	for _, u := range users {
		adminUserResponds = append(adminUserResponds, NewAdminUserRespond(u))
	}
	respond := &AdminUpdateEntityRespond{
		ID:                                 entity.ID.Hex(),
		AccountNumber:                      entity.AccountNumber,
		Name:                               entity.Name,
		Email:                              entity.Email,
		Telephone:                          entity.Telephone,
		IncType:                            entity.IncType,
		CompanyNumber:                      entity.CompanyNumber,
		Website:                            entity.Website,
		DeclaredTurnover:                   entity.DeclaredTurnover,
		Description:                        entity.Description,
		Address:                            entity.Address,
		City:                               entity.City,
		Region:                             entity.Region,
		PostalCode:                         entity.PostalCode,
		Country:                            entity.Country,
		Status:                             entity.Status,
		Offers:                             TagFieldToNames(entity.Offers),
		Wants:                              TagFieldToNames(entity.Wants),
		Categories:                         entity.Categories,
		ShowTagsMatchedSinceLastLogin:      util.ToBool(entity.ShowTagsMatchedSinceLastLogin),
		ReceiveDailyMatchNotificationEmail: util.ToBool(entity.ReceiveDailyMatchNotificationEmail),
		MaxPositiveBalance:                 balanceLimit.MaxPosBal,
		MaxNegativeBalance:                 balanceLimit.MaxNegBal,
		Users:                              adminUserResponds,
		BalanceLimit:                       balanceLimit,
	}
	return respond
}

type AdminUpdateEntityRespond struct {
	ID                                 string              `json:"id"`
	AccountNumber                      string              `json:"accountNumber"`
	Name                               string              `json:"name"`
	Email                              string              `json:"email,omitempty"`
	Telephone                          string              `json:"telephone"`
	IncType                            string              `json:"incType"`
	CompanyNumber                      string              `json:"companyNumber"`
	Website                            string              `json:"website"`
	DeclaredTurnover                   *int                `json:"declaredTurnover"`
	Description                        string              `json:"description"`
	Address                            string              `json:"address"`
	City                               string              `json:"city"`
	Region                             string              `json:"region"`
	PostalCode                         string              `json:"postalCode"`
	Country                            string              `json:"country"`
	Status                             string              `json:"status"`
	Offers                             []string            `json:"offers,omitempty"`
	Wants                              []string            `json:"wants,omitempty"`
	Categories                         []string            `json:"categories,omitempty"`
	ShowTagsMatchedSinceLastLogin      bool                `json:"showTagsMatchedSinceLastLogin"`
	ReceiveDailyMatchNotificationEmail bool                `json:"receiveDailyMatchNotificationEmail"`
	MaxPositiveBalance                 float64             `json:"maxPositiveBalance"`
	MaxNegativeBalance                 float64             `json:"maxNegativeBalance"`
	Users                              []*AdminUserRespond `json:"users"`
	// To log user action.
	BalanceLimit *BalanceLimit `json:"-"`
}

// DELETE /admin/entities/{entityID}

func NewAdminDeleteEntityRespond(entity *Entity) *AdminDeleteEntityRespond {
	return &AdminDeleteEntityRespond{
		ID:                                 entity.ID.Hex(),
		AccountNumber:                      entity.AccountNumber,
		Name:                               entity.Name,
		Email:                              entity.Email,
		Telephone:                          entity.Telephone,
		IncType:                            entity.IncType,
		CompanyNumber:                      entity.CompanyNumber,
		Website:                            entity.Website,
		DeclaredTurnover:                   entity.DeclaredTurnover,
		Description:                        entity.Description,
		Address:                            entity.Address,
		City:                               entity.City,
		Region:                             entity.Region,
		PostalCode:                         entity.PostalCode,
		Country:                            entity.Country,
		Status:                             entity.Status,
		Offers:                             TagFieldToNames(entity.Offers),
		Wants:                              TagFieldToNames(entity.Wants),
		Categories:                         entity.Categories,
		ShowTagsMatchedSinceLastLogin:      util.ToBool(entity.ShowTagsMatchedSinceLastLogin),
		ReceiveDailyMatchNotificationEmail: util.ToBool(entity.ReceiveDailyMatchNotificationEmail),
	}
}

type AdminDeleteEntityRespond struct {
	ID                                 string   `json:"id"`
	AccountNumber                      string   `json:"accountNumber"`
	Name                               string   `json:"name"`
	Email                              string   `json:"email,omitempty"`
	Telephone                          string   `json:"telephone"`
	IncType                            string   `json:"incType"`
	CompanyNumber                      string   `json:"companyNumber"`
	Website                            string   `json:"website"`
	DeclaredTurnover                   *int     `json:"declaredTurnover"`
	Description                        string   `json:"description"`
	Address                            string   `json:"address"`
	City                               string   `json:"city"`
	Region                             string   `json:"region"`
	PostalCode                         string   `json:"postalCode"`
	Country                            string   `json:"country"`
	Status                             string   `json:"status"`
	Offers                             []string `json:"offers,omitempty"`
	Wants                              []string `json:"wants,omitempty"`
	Categories                         []string `json:"categories,omitempty"`
	ShowTagsMatchedSinceLastLogin      bool     `json:"showTagsMatchedSinceLastLogin"`
	ReceiveDailyMatchNotificationEmail bool     `json:"receiveDailyMatchNotificationEmail"`
}

// admin/transfer

type AdminTransferRespond struct {
	TransferID        string     `json:"id"`
	FromAccountNumber string     `json:"fromAccountNumber"`
	FromEntityName    string     `json:"fromEntityName"`
	ToAccountNumber   string     `json:"toAccountNumber"`
	ToEntityName      string     `json:"toEntityName"`
	Amount            float64    `json:"amount"`
	Description       string     `json:"description"`
	Status            string     `json:"status"`
	Type              string     `json:"type,omitempty"`
	CreatedAt         *time.Time `json:"dateProposed,omitempty"`
	CompletedAt       *time.Time `json:"dateCompleted,omitempty"`
}

// GET /admin/transfer
// GET /admin/entities/{entityID}

func NewJournalsToAdminTransfersRespond(journals []*Journal) []*AdminTransferRespond {
	adminTransferRespond := []*AdminTransferRespond{}

	for _, j := range journals {
		t := &AdminTransferRespond{
			TransferID:        j.TransferID,
			FromAccountNumber: j.FromAccountNumber,
			FromEntityName:    j.FromEntityName,
			ToAccountNumber:   j.ToAccountNumber,
			ToEntityName:      j.ToEntityName,
			Amount:            j.Amount,
			Description:       j.Description,
			Type:              j.Type,
			Status:            j.Status,
			CreatedAt:         &j.CreatedAt,
		}
		if j.Status == constant.Transfer.Completed {
			t.CompletedAt = &j.UpdatedAt
		}

		adminTransferRespond = append(adminTransferRespond, t)
	}

	return adminTransferRespond
}

type AdminSearchTransferRespond struct {
	Transfers       []*AdminTransferRespond
	NumberOfResults int
	TotalPages      int
}

// POST /admin/transfers
// GET /admin/transfers/{transferID}

func NewJournalToAdminTransferRespond(j *Journal) *AdminTransferRespond {
	res := &AdminTransferRespond{
		TransferID:        j.TransferID,
		FromAccountNumber: j.FromAccountNumber,
		FromEntityName:    j.FromEntityName,
		ToAccountNumber:   j.ToAccountNumber,
		ToEntityName:      j.ToEntityName,
		Amount:            j.Amount,
		Description:       j.Description,
		Status:            j.Status,
		Type:              j.Type,
		CreatedAt:         &j.CreatedAt,
	}
	if j.Status == constant.Transfer.Completed {
		res.CompletedAt = &j.UpdatedAt
	}
	return res
}
