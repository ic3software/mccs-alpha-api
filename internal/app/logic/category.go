package logic

import (
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/mongo"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type category struct{}

var Category = &category{}

func (c *category) Search(req *types.SearchCategoryReqBody) (*types.FindCategoryResult, error) {
	result, err := mongo.Category.Search(req)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *category) FindByIDString(id string) (*types.Category, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	category, err := mongo.Category.FindByID(objectID)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (c *category) FindByName(name string) (*types.Category, error) {
	category, err := mongo.Category.FindByName(name)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (c *category) CreateOne(categoryName string) (*types.Category, error) {
	created, err := mongo.Category.Create(categoryName)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (c *category) Create(categories ...string) error {
	if len(categories) == 1 {
		_, err := mongo.Category.Create(categories[0])
		if err != nil {
			return err
		}
		return nil
	}
	for _, category := range categories {
		_, err := mongo.Category.Create(category)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *category) FindOneAndUpdate(id primitive.ObjectID, update *types.Category) (*types.Category, error) {
	updated, err := mongo.Category.FindOneAndUpdate(id, update)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (c *category) FindOneAndDelete(id primitive.ObjectID) (*types.Category, error) {
	deleted, err := mongo.Category.FindOneAndDelete(id)
	if err != nil {
		return nil, err
	}
	return deleted, nil
}

// TO BE REMOVED

func (c *category) FindTags(name string, page int64) (*types.FindCategoryResult, error) {
	result, err := mongo.Category.FindTags(name, page)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *category) GetAll() ([]*types.Category, error) {
	categories, err := mongo.Category.GetAll()
	if err != nil {
		return nil, err
	}
	return categories, nil
}
