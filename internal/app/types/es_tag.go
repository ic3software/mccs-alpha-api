package types

import "time"

type TagESRecord struct {
	TagID        string    `json:"tagID,omitempty"`
	Name         string    `json:"name,omitempty"`
	OfferAddedAt time.Time `json:"offerAddedAt,omitempty"`
	WantAddedAt  time.Time `json:"wantAddedAt,omitempty"`
}
