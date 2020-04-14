package logic

import (
	"time"

	"github.com/ic3network/mccs-alpha-api/internal/app/repository/es"
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/mongo"
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/pg"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type entity struct{}

var Entity = &entity{}

func (_ *entity) Create(entity *types.Entity) (primitive.ObjectID, error) {
	account, err := pg.Account.Create()
	if err != nil {
		return primitive.ObjectID{}, err
	}
	entity.AccountNumber = account.AccountNumber
	id, err := mongo.Entity.Create(entity)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	err = es.Entity.Create(id, entity)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return id, nil
}

func (_ *entity) AssociateUser(entityID, userID primitive.ObjectID) error {
	err := mongo.Entity.AssociateUser(entityID, userID)
	if err != nil {
		return err
	}
	return nil
}

func (_ *entity) FindByID(objectID primitive.ObjectID) (*types.Entity, error) {
	entity, err := mongo.Entity.FindByID(objectID)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (_ *entity) FindByIDs(ids []primitive.ObjectID) ([]*types.Entity, error) {
	entity, err := mongo.Entity.FindByIDs(ids)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (_ *entity) FindByAccountNumber(accountNumber string) (*types.Entity, error) {
	entity, err := mongo.Entity.FindByAccountNumber(accountNumber)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (_ *entity) FindByStringID(id string) (*types.Entity, error) {
	objectID, _ := primitive.ObjectIDFromHex(id)
	entity, err := mongo.Entity.FindByID(objectID)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (_ *entity) FindOneAndUpdate(update *types.Entity) (*types.Entity, error) {
	err := es.Entity.Update(update)
	if err != nil {
		return nil, err
	}
	entity, err := mongo.Entity.FindOneAndUpdate(update)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (_ *entity) AdminFindOneAndUpdate(update *types.Entity) (*types.Entity, error) {
	err := es.Entity.AdminUpdate(update)
	if err != nil {
		return nil, err
	}
	entity, err := mongo.Entity.FindOneAndUpdate(update)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (_ *entity) AdminFindOneAndDelete(id primitive.ObjectID) (*types.Entity, error) {
	err := es.Entity.Delete(id.Hex())
	if err != nil {
		return nil, err
	}
	entity, err := mongo.Entity.AdminFindOneAndDelete(id)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (_ *entity) UpdateTags(id primitive.ObjectID, difference *types.TagDifference) error {
	err := mongo.Entity.UpdateTags(id, difference)
	if err != nil {
		return err
	}
	err = es.Entity.UpdateTags(id, difference)
	if err != nil {
		return err
	}
	return nil
}

func (_ *entity) Search(req *types.SearchEntityReqBody) (*types.SearchEntityResult, error) {
	result, err := es.Entity.Search(req)
	if err != nil {
		return nil, err
	}
	entities, err := mongo.Entity.FindByStringIDs(result.IDs)
	if err != nil {
		return nil, err
	}
	return &types.SearchEntityResult{
		Entities:        entities,
		NumberOfResults: result.NumberOfResults,
		TotalPages:      result.TotalPages,
	}, nil
}

func (_ *entity) AdminSearch(req *types.AdminSearchEntityReqBody) (*types.SearchEntityResult, error) {
	result, err := es.Entity.AdminSearch(req)
	if err != nil {
		return nil, err
	}
	entities, err := mongo.Entity.FindByStringIDs(result.IDs)
	if err != nil {
		return nil, err
	}
	return &types.SearchEntityResult{
		Entities:        entities,
		NumberOfResults: result.NumberOfResults,
		TotalPages:      result.TotalPages,
	}, nil
}

func (_ *entity) AddToFavoriteEntities(req *types.AddToFavoriteReqBody) error {
	err := mongo.Entity.AddToFavoriteEntities(req)
	if err != nil {
		return err
	}
	return nil
}

func (b *entity) UpdateAllTagsCreatedAt(id primitive.ObjectID, t time.Time) error {
	err := es.Entity.UpdateAllTagsCreatedAt(id, t)
	if err != nil {
		return err
	}
	err = mongo.Entity.UpdateAllTagsCreatedAt(id, t)
	if err != nil {
		return err
	}
	return nil
}

func (b *entity) SetMemberStartedAt(id primitive.ObjectID) error {
	err := mongo.Entity.SetMemberStartedAt(id)
	if err != nil {
		return err
	}
	return nil
}

func (b *entity) RenameCategory(old string, new string) error {
	err := es.Entity.RenameCategory(old, new)
	if err != nil {
		return err
	}
	err = mongo.Entity.RenameCategory(old, new)
	if err != nil {
		return err
	}
	return nil
}

func (b *entity) DeleteCategory(name string) error {
	err := es.Entity.DeleteCategory(name)
	if err != nil {
		return err
	}
	err = mongo.Entity.DeleteCategory(name)
	if err != nil {
		return err
	}
	return nil
}

func (b *entity) RenameTag(old string, new string) error {
	err := es.Entity.RenameTag(old, new)
	if err != nil {
		return err
	}
	err = mongo.Entity.RenameTag(old, new)
	if err != nil {
		return err
	}
	return nil
}

func (b *entity) DeleteTag(name string) error {
	err := es.Entity.DeleteTag(name)
	if err != nil {
		return err
	}
	err = mongo.Entity.DeleteTag(name)
	if err != nil {
		return err
	}
	return nil
}
