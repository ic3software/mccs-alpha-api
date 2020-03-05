package logic

import (
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/mongo"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type category struct{}

var Category = &category{}

func (c *category) Create(categories ...string) error {
	if len(categories) == 1 {
		err := mongo.Category.Create(categories[0])
		if err != nil {
			return err
		}
		return nil
	}
	for _, category := range categories {
		err := mongo.Category.Create(category)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *category) Find(query *types.SearchCategoryQuery) (*types.FindCategoryResult, error) {
	result, err := mongo.Category.Find(query)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// TO BE REMOVED

func (c *category) FindByName(name string) (*types.Category, error) {
	adminTag, err := mongo.Category.FindByName(name)
	if err != nil {
		return nil, err
	}
	return adminTag, nil
}

func (c *category) FindByID(id primitive.ObjectID) (*types.Category, error) {
	adminTag, err := mongo.Category.FindByID(id)
	if err != nil {
		return nil, err
	}
	return adminTag, nil
}

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

func (c *category) Update(tag *types.Category) error {
	err := mongo.Category.Update(tag)
	if err != nil {
		return err
	}
	return nil
}

func (c *category) DeleteByID(id primitive.ObjectID) error {
	err := mongo.Category.DeleteByID(id)
	if err != nil {
		return err
	}
	return nil
}
