package types

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Entity is the model representation of a entity in the data model.
type Entity struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	DeletedAt time.Time          `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`

	Users []primitive.ObjectID `json:"users,omitempty" bson:"users,omitempty"`

	EntityName         string      `json:"entityName,omitempty" bson:"entityName,omitempty"`
	EntityPhone        string      `json:"entityPhone,omitempty" bson:"entityPhone,omitempty"`
	Email              string      `json:"email,omitempty" bson:"email,omitempty"`
	IncType            string      `json:"incType,omitempty" bson:"incType,omitempty"`
	CompanyNumber      string      `json:"companyNumber,omitempty" bson:"companyNumber,omitempty"`
	Website            string      `json:"website,omitempty" bson:"website,omitempty"`
	Turnover           int         `json:"turnover,omitempty" bson:"turnover,omitempty"`
	Offers             []*TagField `json:"offers,omitempty" bson:"offers,omitempty"`
	Wants              []*TagField `json:"wants,omitempty" bson:"wants,omitempty"`
	Description        string      `json:"description,omitempty" bson:"description,omitempty"`
	LocationAddress    string      `json:"locationAddress,omitempty" bson:"locationAddress,omitempty"`
	LocationCity       string      `json:"locationCity,omitempty" bson:"locationCity,omitempty"`
	LocationRegion     string      `json:"locationRegion,omitempty" bson:"locationRegion,omitempty"`
	LocationPostalCode string      `json:"locationPostalCode,omitempty" bson:"locationPostalCode,omitempty"`
	LocationCountry    string      `json:"locationCountry,omitempty" bson:"locationCountry,omitempty"`
	Status             string      `json:"status,omitempty" bson:"status,omitempty"`
	Categories         []string    `json:"categories,omitempty" bson:"categories,omitempty"`
	// Timestamp when trading status applied
	MemberStartedAt time.Time `json:"memberStartedAt,omitempty" bson:"memberStartedAt,omitempty"`

	AccountNumber    string               `json:"accountNumber,omitempty" bson:"accountNumber,omitempty"`
	FavoriteEntities []primitive.ObjectID `json:"favoriteEntities,omitempty" bson:"favoriteEntities,omitempty"`
}

func (entity *Entity) Validate() []error {
	errs := []error{}
	if len(entity.EntityName) > 100 {
		errs = append(errs, errors.New("Entity name length cannot exceed 100 characters."))
	}
	if len(entity.EntityPhone) > 25 {
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
	if len(entity.Description) > 500 {
		errs = append(errs, errors.New("Description length cannot exceed 500 characters."))
	}
	if len(entity.LocationCountry) > 10 {
		errs = append(errs, errors.New("Country length cannot exceed 50 characters."))
	}
	if len(entity.LocationCity) > 10 {
		errs = append(errs, errors.New("City length cannot exceed 50 characters."))
	}
	if len(entity.LocationAddress) > 255 {
		errs = append(errs, errors.New("Address length cannot exceed 255 characters."))
	}
	if len(entity.LocationRegion) > 50 {
		errs = append(errs, errors.New("Region length cannot exceed 50 characters."))
	}
	if len(entity.LocationPostalCode) > 10 {
		errs = append(errs, errors.New("Postal code length cannot exceed 10 characters."))
	}
	return errs
}

// Helper types

type TagDifference struct {
	OffersAdded   []string
	OffersRemoved []string
	WantsAdded    []string
	WantsRemoved  []string
}

type FindEntityResult struct {
	Entities        []*Entity
	NumberOfResults int
	TotalPages      int
}

// TO BE REMOVED

type EntityData struct {
	ID                 primitive.ObjectID
	EntityName         string
	IncType            string
	CompanyNumber      string
	EntityPhone        string
	Website            string
	Turnover           int
	Offers             []*TagField
	Wants              []*TagField
	OffersAdded        []string
	OffersRemoved      []string
	WantsAdded         []string
	WantsRemoved       []string
	Description        string
	LocationAddress    string
	LocationCity       string
	LocationRegion     string
	LocationPostalCode string
	LocationCountry    string
	Status             string
	Categories         []string
}
