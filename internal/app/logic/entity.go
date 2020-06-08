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

func (_ *entity) Create(entity *types.Entity) (*types.Entity, error) {
	account, err := pg.Account.Create()
	if err != nil {
		return nil, err
	}
	entity.AccountNumber = account.AccountNumber
	created, err := mongo.Entity.Create(entity)
	if err != nil {
		return nil, err
	}
	err = es.Entity.Create(created.ID, entity)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// POST /signup

func (_ *entity) AssociateUser(entityID, userID primitive.ObjectID) error {
	err := mongo.Entity.AssociateUser([]primitive.ObjectID{entityID}, userID)
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

func (_ *entity) EmailExists(email string) bool {
	_, err := mongo.Entity.FindByEmail(email)
	if err != nil {
		return false
	}
	return true
}

// PATCH /user/entities/{entityID}

func (_ *entity) FindOneAndUpdate(req *types.UpdateUserEntityReq) (*types.Entity, error) {
	err := es.Entity.Update(req)
	if err != nil {
		return nil, err
	}
	entity, err := mongo.Entity.FindOneAndUpdate(req)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

// PATCH /admin/entities/{entityID}

func (_ *entity) AdminFindOneAndUpdate(req *types.AdminUpdateEntityReq) (*types.Entity, error) {
	err := es.Entity.AdminUpdate(req)
	if err != nil {
		return nil, err
	}
	err = pg.BalanceLimit.AdminUpdate(req)
	if err != nil {
		return nil, err
	}
	entity, err := mongo.Entity.AdminFindOneAndUpdate(req)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

// DELETE /admin/entities/{entityID}

func (_ *entity) AdminFindOneAndDelete(id primitive.ObjectID) (*types.Entity, error) {
	err := es.Entity.Delete(id.Hex())
	if err != nil {
		return nil, err
	}
	deleted, err := mongo.Entity.AdminFindOneAndDelete(id)
	if err != nil {
		return nil, err
	}
	err = pg.Account.Delete(deleted.AccountNumber)
	if err != nil {
		return nil, err
	}
	return deleted, nil
}

func (_ *entity) Search(req *types.SearchEntityReq) (*types.SearchEntityResult, error) {
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

func (_ *entity) AdminSearch(req *types.AdminSearchEntityReq) (*types.SearchEntityResult, error) {
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

func (_ *entity) AddToFavoriteEntities(req *types.AddToFavoriteReq) error {
	err := mongo.Entity.AddToFavoriteEntities(req)
	if err != nil {
		return err
	}
	return nil
}

func (e *entity) UpdateAllTagsCreatedAt(id primitive.ObjectID, t time.Time) error {
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

func (e *entity) SetMemberStartedAt(id primitive.ObjectID) error {
	err := mongo.Entity.SetMemberStartedAt(id)
	if err != nil {
		return err
	}
	return nil
}

func (e *entity) RenameCategory(old string, new string) error {
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

func (e *entity) DeleteCategory(name string) error {
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

func (e *entity) RenameTag(old string, new string) error {
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

func (e *entity) DeleteTag(name string) error {
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

// daily_email_schedule

func (e *entity) FindByDailyNotification() ([]*types.Entity, error) {
	entities, err := mongo.Entity.FindByDailyNotification()
	if err != nil {
		return nil, err
	}
	return entities, nil
}

func (e *entity) UpdateLastNotificationSentDate(id primitive.ObjectID) error {
	err := mongo.Entity.UpdateLastNotificationSentDate(id)
	if err != nil {
		return err
	}
	return nil
}
