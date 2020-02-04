package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Entity is the model representation of a entity in the data model.
type Entity struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	DeletedAt time.Time          `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`

	EntityName         string      `json:"entityName,omitempty" bson:"entityName,omitempty"`
	EntityPhone        string      `json:"entityPhone,omitempty" bson:"entityPhone,omitempty"`
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
	AdminTags          []string    `json:"adminTags,omitempty" bson:"adminTags,omitempty"`
	// Timestamp when trading status applied
	MemberStartedAt time.Time `json:"memberStartedAt,omitempty" bson:"memberStartedAt,omitempty"`
}

type TagField struct {
	Name      string    `json:"name,omitempty" bson:"name,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}

// EntityESRecord is the data that will store into the elastic search.
type EntityESRecord struct {
	EntityID        string      `json:"entityID,omitempty"`
	EntityName      string      `json:"entityName,omitempty"`
	Offers          []*TagField `json:"offers,omitempty"`
	Wants           []*TagField `json:"wants,omitempty"`
	LocationCity    string      `json:"locationCity,omitempty"`
	LocationCountry string      `json:"locationCountry,omitempty"`
	Status          string      `json:"status,omitempty"`
	AdminTags       []string    `json:"adminTags,omitempty"`
}

// Helper types

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
	AdminTags          []string
}

type SearchCriteria struct {
	TagType          string
	Tags             []*TagField
	CreatedOnOrAfter time.Time

	Statuses              []string // accepted", "pending", rejected", "tradingPending", "tradingAccepted", "tradingRejected"
	EntityName            string
	LocationCountry       string
	LocationCity          string
	ShowUserFavoritesOnly bool
	FavoriteEntities      []primitive.ObjectID
	AdminTag              string
}

type FindEntityResult struct {
	Entities        []*Entity
	NumberOfResults int
	TotalPages      int
}
