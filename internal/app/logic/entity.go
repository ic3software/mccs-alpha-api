package logic

import (
	"time"

	"github.com/ic3network/mccs-alpha-api/internal/app/repository/es"
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/mongo"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/e"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type entity struct{}

var Entity = &entity{}

func (_ *entity) Create(entity *types.Entity) (primitive.ObjectID, error) {
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

// OLD CODE

func (b *entity) FindByID(id primitive.ObjectID) (*types.Entity, error) {
	bs, err := mongo.Entity.FindByID(id)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func (b *entity) UpdateEntity(
	id primitive.ObjectID,
	entity *types.EntityData,
	isAdmin bool,
) error {
	err := es.Entity.UpdateEntity(id, entity)
	if err != nil {
		return e.Wrap(err, "update entity failed")
	}
	err = mongo.Entity.UpdateEntity(id, entity, isAdmin)
	if err != nil {
		return e.Wrap(err, "update entity failed")
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

func (b *entity) UpdateAllTagsCreatedAt(id primitive.ObjectID, t time.Time) error {
	err := es.Entity.UpdateAllTagsCreatedAt(id, t)
	if err != nil {
		return e.Wrap(err, "EntityService UpdateAllTagsCreatedAt failed")
	}
	err = mongo.Entity.UpdateAllTagsCreatedAt(id, t)
	if err != nil {
		return e.Wrap(err, "EntityService UpdateAllTagsCreatedAt failed")
	}
	return nil
}

func (b *entity) FindEntity(c *types.SearchCriteria, page int64) (*types.FindEntityResult, error) {
	ids, numberOfResults, totalPages, err := es.Entity.Find(c, page)
	if err != nil {
		return nil, e.Wrap(err, "EntityService FindEntity failed")
	}
	entities, err := mongo.Entity.FindByIDs(ids)
	if err != nil {
		return nil, e.Wrap(err, "EntityService FindEntity failed")
	}
	return &types.FindEntityResult{
		Entities:        entities,
		NumberOfResults: numberOfResults,
		TotalPages:      totalPages,
	}, nil
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
