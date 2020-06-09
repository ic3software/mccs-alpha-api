package types

import (
	"errors"
	"time"

	"github.com/ic3network/mccs-alpha-api/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Entity is the model representation of a entity in the data model.
type Entity struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	DeletedAt time.Time          `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`

	Users []primitive.ObjectID `json:"users,omitempty" bson:"users,omitempty"`

	Name             string      `json:"name,omitempty" bson:"name,omitempty"`
	Telephone        string      `json:"telephone,omitempty" bson:"telephone,omitempty"`
	Email            string      `json:"email,omitempty" bson:"email,omitempty"`
	IncType          string      `json:"incType,omitempty" bson:"incType,omitempty"`
	CompanyNumber    string      `json:"companyNumber,omitempty" bson:"companyNumber,omitempty"`
	Website          string      `json:"website,omitempty" bson:"website,omitempty"`
	DeclaredTurnover *int        `json:"declaredTurnover,omitempty" bson:"declaredTurnover,omitempty"`
	Offers           []*TagField `json:"offers,omitempty" bson:"offers,omitempty"`
	Wants            []*TagField `json:"wants,omitempty" bson:"wants,omitempty"`
	Description      string      `json:"description,omitempty" bson:"description,omitempty"`
	Address          string      `json:"address,omitempty" bson:"address,omitempty"`
	City             string      `json:"city,omitempty" bson:"city,omitempty"`
	Region           string      `json:"region,omitempty" bson:"region,omitempty"`
	PostalCode       string      `json:"postalCode,omitempty" bson:"postalCode,omitempty"`
	Country          string      `json:"country,omitempty" bson:"country,omitempty"`
	Status           string      `json:"status,omitempty" bson:"status,omitempty"`
	Categories       []string    `json:"categories,omitempty" bson:"categories,omitempty"`
	// Timestamp when trading status applied
	MemberStartedAt time.Time `json:"memberStartedAt,omitempty" bson:"memberStartedAt,omitempty"`

	// flags
	ShowTagsMatchedSinceLastLogin      *bool `json:"showTagsMatchedSinceLastLogin,omitempty" bson:"showTagsMatchedSinceLastLogin,omitempty"`
	ReceiveDailyMatchNotificationEmail *bool `json:"receiveDailyMatchNotificationEmail,omitempty" bson:"receiveDailyMatchNotificationEmail,omitempty"`

	LastNotificationSentDate time.Time `json:"lastNotificationSentDate,omitempty" bson:"lastNotificationSentDate,omitempty"`

	AccountNumber    string               `json:"accountNumber,omitempty" bson:"accountNumber,omitempty"`
	FavoriteEntities []primitive.ObjectID `json:"favoriteEntities,omitempty" bson:"favoriteEntities,omitempty"`
}

func (entity *Entity) Validate() []error {
	errs := []error{}

	if len(entity.Email) != 0 {
		errs = append(errs, util.ValidateEmail(entity.Email)...)
	}
	if len(entity.Status) != 0 && !util.IsValidStatus(entity.Status) {
		errs = append(errs, errors.New("Please specify a valid status."))
	}

	if len(entity.Name) > 100 {
		errs = append(errs, errors.New("Entity name length cannot exceed 100 characters."))
	}
	if len(entity.Telephone) > 25 {
		errs = append(errs, errors.New("Telephone length cannot exceed 25 characters."))
	}
	if len(entity.IncType) > 25 {
		errs = append(errs, errors.New("Incorporation type length cannot exceed 25 characters."))
	}
	if len(entity.CompanyNumber) > 20 {
		errs = append(errs, errors.New("Company number length cannot exceed 20 characters."))
	}
	if len(entity.Website) > 100 {
		errs = append(errs, errors.New("Website URL length cannot exceed 100 characters."))
	}
	if entity.DeclaredTurnover != nil && *entity.DeclaredTurnover < 0 {
		errs = append(errs, errors.New("Declared turnover should be a positive number."))
	}
	if len(entity.Description) > 500 {
		errs = append(errs, errors.New("Description length cannot exceed 500 characters."))
	}
	if len(entity.Address) > 255 {
		errs = append(errs, errors.New("Address length cannot exceed 255 characters."))
	}
	if len(entity.City) > 50 {
		errs = append(errs, errors.New("City length cannot exceed 50 characters."))
	}
	if len(entity.Region) > 50 {
		errs = append(errs, errors.New("Region length cannot exceed 50 characters."))
	}
	if len(entity.PostalCode) > 10 {
		errs = append(errs, errors.New("Postal code length cannot exceed 10 characters."))
	}
	if len(entity.Country) > 50 {
		errs = append(errs, errors.New("Country length cannot exceed 50 characters."))
	}
	return errs
}

// Helper types

type SearchEntityResult struct {
	Entities        []*Entity
	NumberOfResults int
	TotalPages      int
}

type UpdateOfferAndWants struct {
	EntityID      primitive.ObjectID
	OriginStatus  string
	UpdatedStatus string
	UpdatedOffers []string
	UpdatedWants  []string
	AddedOffers   []string
	AddedWants    []string
}
