package logic

import (
	"time"

	"github.com/ic3network/mccs-alpha-api/internal/app/repository/es"
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/mongo"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
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

func (t *tag) Search(req *types.SearchTagReq) (*types.FindTagResult, error) {
	found, err := mongo.Tag.Search(req)
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

func (t *tag) FindByID(objectID primitive.ObjectID) (*types.Tag, error) {
	tag, err := mongo.Tag.FindByID(objectID)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

func (t *tag) FindByIDString(id string) (*types.Tag, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	tag, err := mongo.Tag.FindByID(objectID)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

func (t *tag) FindOneAndUpdate(id primitive.ObjectID, update *types.Tag) (*types.Tag, error) {
	err := es.Tag.Update(id, update)
	if err != nil {
		return nil, err
	}
	updated, err := mongo.Tag.FindOneAndUpdate(id, update)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (t *tag) FindOneAndDelete(id primitive.ObjectID) (*types.Tag, error) {
	err := es.Tag.DeleteByID(id.Hex())
	if err != nil {
		return nil, err
	}
	deleted, err := mongo.Tag.FindOneAndDelete(id)
	if err != nil {
		return nil, err
	}
	return deleted, nil
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

// MatchOffers loops through user's offers and finds out the matched wants.
// Only add to the result when matches more than one tag.
func (t *tag) MatchOffers(offers []string, lastLoginDate time.Time) (map[string][]string, error) {
	resultMap := make(map[string][]string, len(offers))

	for _, offer := range offers {
		matches, err := es.Tag.MatchOffer(offer, lastLoginDate)
		if err != nil {
			return nil, err
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
			return nil, err
		}
		if len(matches) > 0 {
			resultMap[want] = matches
		}
	}

	return resultMap, nil
}
