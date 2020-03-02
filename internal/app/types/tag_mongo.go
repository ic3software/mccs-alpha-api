package types

import (
	"time"

	"github.com/ic3network/mccs-alpha-api/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Tag is the model representation of a tag in the data model.
type Tag struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	DeletedAt time.Time          `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`

	Name         string    `json:"name,omitempty" bson:"name,omitempty"`
	OfferAddedAt time.Time `json:"offerAddedAt,omitempty" bson:"offerAddedAt,omitempty"`
	WantAddedAt  time.Time `json:"wantAddedAt,omitempty" bson:"wantAddedAt,omitempty"`
}

type TagESRecord struct {
	TagID        string    `json:"tagID,omitempty"`
	Name         string    `json:"name,omitempty"`
	OfferAddedAt time.Time `json:"offerAddedAt,omitempty"`
	WantAddedAt  time.Time `json:"wantAddedAt,omitempty"`
}

type TagDifference struct {
	OffersAdded   []string
	OffersRemoved []string
	WantsAdded    []string
	WantsRemoved  []string
}

func NewTagDifference(oldOffers, newOffers, oldWants, newWants []string) *TagDifference {
	var offersAdded, offersRemoved, wantsAdded, wantsRemoved []string
	if len(newOffers) != 0 {
		offersAdded, offersRemoved = util.TagDifference(newOffers, oldOffers)
	}
	if len(newWants) != 0 {
		wantsAdded, wantsRemoved = util.TagDifference(newWants, oldWants)
	}
	return &TagDifference{
		OffersAdded:   offersAdded,
		OffersRemoved: offersRemoved,
		WantsAdded:    wantsAdded,
		WantsRemoved:  wantsRemoved,
	}
}

func TagToNames(tags []*Tag) []string {
	names := make([]string, 0, len(tags))
	for _, t := range tags {
		names = append(names, t.Name)
	}
	return names
}

// Helper types

type FindTagResult struct {
	Tags            []*Tag
	NumberOfResults int
	TotalPages      int
}

type MatchedTags struct {
	MatchedOffers map[string][]string
	MatchedWants  map[string][]string
}
