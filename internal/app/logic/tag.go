package logic

import (
	"time"

	"github.com/ic3network/mccs-alpha-api/internal/app/repository/es"
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/mongo"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/e"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type tag struct{}

var Tag = &tag{}

func (t *tag) Create(name string) (*types.Tag, error) {
	created, err := mongo.Tag.Create(name)
	if err != nil {
		return nil, err
	}
	err = es.Tag.Create(created.ID, name)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (t *tag) Search(query *types.SearchTagQuery) (*types.FindTagResult, error) {
	found, err := mongo.Tag.Search(query)
	if err != nil {
		return nil, err
	}
	return found, nil
}

func (t *tag) FindByName(name string) (*types.Tag, error) {
	tag, err := mongo.Tag.FindByName(name)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

// UpdateOffer will add/modify the offer tag.
func (t *tag) UpdateOffer(name string) error {
	id, err := mongo.Tag.UpdateOffer(name)
	if err != nil {
		return err
	}
	err = es.Tag.UpdateOffer(id.Hex(), name)
	if err != nil {
		return err
	}
	return nil
}

// UpdateWant will add/modify the want tag.
func (t *tag) UpdateWant(name string) error {
	id, err := mongo.Tag.UpdateWant(name)
	if err != nil {
		return err
	}
	err = es.Tag.UpdateWant(id.Hex(), name)
	if err != nil {
		return err
	}
	return nil
}

// TO BE REMOVED

func (t *tag) FindByID(id primitive.ObjectID) (*types.Tag, error) {
	tag, err := mongo.Tag.FindByID(id)
	if err != nil {
		return nil, e.Wrap(err, "TagService FindByID failed")
	}
	return tag, nil
}

func (t *tag) Rename(tag *types.Tag) error {
	err := es.Tag.Rename(tag)
	if err != nil {
		return e.Wrap(err, "TagService Rename failed")
	}
	err = mongo.Tag.Rename(tag)
	if err != nil {
		return e.Wrap(err, "TagService Rename failed")
	}
	return nil
}

func (t *tag) DeleteByID(id primitive.ObjectID) error {
	err := es.Tag.DeleteByID(id.Hex())
	if err != nil {
		return e.Wrap(err, "TagService DeleteByID failed")
	}
	err = mongo.Tag.DeleteByID(id)
	if err != nil {
		return e.Wrap(err, "TagService DeleteByID failed")
	}
	return nil
}

// MatchOffers loops through user's offers and finds out the matched wants.
// Only add to the result when matches more than one tag.
func (t *tag) MatchOffers(offers []string, lastLoginDate time.Time) (map[string][]string, error) {
	resultMap := make(map[string][]string, len(offers))

	for _, offer := range offers {
		matches, err := es.Tag.MatchOffer(offer, lastLoginDate)
		if err != nil {
			return nil, e.Wrap(err, "TagService MatchOffers failed")
		}
		if len(matches) > 0 {
			resultMap[offer] = matches
		}
	}

	return resultMap, nil
}

// MatchWants loops through user's wants and finds out the matched offers.
// Only add to the result when matches more than one tag.
func (t *tag) MatchWants(wants []string, lastLoginDate time.Time) (map[string][]string, error) {
	resultMap := make(map[string][]string, len(wants))

	for _, want := range wants {
		matches, err := es.Tag.MatchWant(want, lastLoginDate)
		if err != nil {
			return nil, e.Wrap(err, "TagService MatchWants failed")
		}
		if len(matches) > 0 {
			resultMap[want] = matches
		}
	}

	return resultMap, nil
}
