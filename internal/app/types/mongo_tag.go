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
	NewAddedOffers []string
	NewAddedWants  []string
	OffersRemoved  []string
	WantsRemoved   []string
	// End result for offers and wants.
	Offers []string
	Wants  []string
}

func NewTagDifference(oldOffers, updateOffers, oldWants, updateWants []string) *TagDifference {
	var offers, wants, newAddedOffers, offersRemoved, newAddedWants, wantsRemoved []string

	if len(updateOffers) != 0 {
		offers = updateOffers
		newAddedOffers, offersRemoved = util.TagDifference(updateOffers, oldOffers)
	} else {
		offers = oldOffers
	}
	if len(updateWants) != 0 {
		wants = updateWants
		newAddedWants, wantsRemoved = util.TagDifference(updateWants, oldWants)
	} else {
		wants = oldWants
	}

	return &TagDifference{
		Offers:         offers,
		Wants:          wants,
		NewAddedOffers: newAddedOffers,
		NewAddedWants:  newAddedWants,
		OffersRemoved:  offersRemoved,
		WantsRemoved:   wantsRemoved,
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
