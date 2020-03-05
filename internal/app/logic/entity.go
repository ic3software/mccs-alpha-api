package logic

import (
	"time"

	"github.com/ic3network/mccs-alpha-api/internal/app/repository/es"
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/mongo"
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/pg"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/e"
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

func (_ *entity) Find(query *types.SearchEntityQuery) (*types.FindEntityResult, error) {
	result, err := es.Entity.Find(query)
	if err != nil {
		return nil, err
	}
	entities, err := mongo.Entity.FindByIDs(result.IDs)
	if err != nil {
		return nil, err
	}
	return &types.FindEntityResult{
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

// TO BE REMOVED

func (b *entity) UpdateEntity(id primitive.ObjectID, difference *types.EntityData, isAdmin bool) error {
	return nil
}

func (b *entity) DeleteByID(id primitive.ObjectID) error {
	err := es.Entity.Delete(id.Hex())
	if err != nil {
		return e.Wrap(err, "delete entity by id failed")
	}
	err = mongo.Entity.DeleteByID(id)
	if err != nil {
		return e.Wrap(err, "delete entity by id failed")
	}
	return nil
}

func (b *entity) RenameTag(old string, new string) error {
	err := es.Entity.RenameTag(old, new)
	if err != nil {
		return e.Wrap(err, "EntityMongo RenameTag failed")
	}
	err = mongo.Entity.RenameTag(old, new)
	if err != nil {
		return e.Wrap(err, "EntityMongo RenameTag failed")
	}
	return nil
}

func (b *entity) RenameAdminTag(old string, new string) error {
	err := es.Entity.RenameAdminTag(old, new)
	if err != nil {
		return e.Wrap(err, "EntityMongo RenameAdminTag failed")
	}
	err = mongo.Entity.RenameAdminTag(old, new)
	if err != nil {
		return e.Wrap(err, "EntityMongo RenameAdminTag failed")
	}
	return nil
}

func (b *entity) DeleteTag(name string) error {
	err := es.Entity.DeleteTag(name)
	if err != nil {
		return e.Wrap(err, "EntityMongo DeleteTag failed")
	}
	err = mongo.Entity.DeleteTag(name)
	if err != nil {
		return e.Wrap(err, "EntityMongo DeleteTag failed")
	}
	return nil
}

func (b *entity) DeleteAdminTags(name string) error {
	err := es.Entity.DeleteAdminTags(name)
	if err != nil {
		return e.Wrap(err, "EntityMongo DeleteAdminTags failed")
	}
	err = mongo.Entity.DeleteAdminTags(name)
	if err != nil {
		return e.Wrap(err, "EntityMongo DeleteAdminTags failed")
	}
	return nil
}
