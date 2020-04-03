package mongo

import (
	"context"
	"strings"
	"time"

	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/e"
	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type category struct {
	c *mongo.Collection
}

var Category = &category{}

func (c *category) Register(db *mongo.Database) {
	c.c = db.Collection("categories")
}

func (c *category) Create(name string) error {
	if name == "" || len(strings.TrimSpace(name)) == 0 {
		return nil
	}

	filter := bson.M{"name": name}
	update := bson.M{"$setOnInsert": bson.M{"name": name, "createdAt": time.Now()}}
	_, err := c.c.UpdateOne(
		context.Background(),
		filter,
		update,
		options.Update().SetUpsert(true),
	)
	return err
}

func (c *category) Search(query *types.SearchCategoryQuery) (*types.FindCategoryResult, error) {
	var results []*types.Category

	findOptions := options.Find()
	findOptions.SetSkip(int64(query.PageSize * (query.Page - 1)))
	findOptions.SetLimit(int64(query.PageSize))

	filter := bson.M{
		"name":      primitive.Regex{Pattern: "^" + query.Prefix + ".*" + query.Fragment + ".*", Options: "i"},
		"deletedAt": bson.M{"$exists": false},
	}
	cur, err := c.c.Find(context.TODO(), filter, findOptions)
	if err != nil {
		return nil, err
	}

	for cur.Next(context.TODO()) {
		var elem types.Category
		err := cur.Decode(&elem)
		if err != nil {
			return nil, err
		}
		results = append(results, &elem)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	cur.Close(context.TODO())

	totalCount, err := c.c.CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	return &types.FindCategoryResult{
		Categories:      results,
		NumberOfResults: int(totalCount),
		TotalPages:      util.GetNumberOfPages(int(totalCount), query.PageSize),
	}, nil
}

func (c *category) FindByName(name string) (*types.Category, error) {
	category := types.Category{}
	filter := bson.M{
		"name":      name,
		"deletedAt": bson.M{"$exists": false},
	}
	err := c.c.FindOne(context.Background(), filter).Decode(&category)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (c *category) FindOneAndUpdate(id primitive.ObjectID, update *types.Category) (*types.Category, error) {
	filter := bson.M{"_id": id}
	result := c.c.FindOneAndUpdate(
		context.Background(),
		filter,
		bson.M{
			"$set": bson.M{
				"name":      update.Name,
				"updatedAt": time.Now(),
			},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	if result.Err() != nil {
		return nil, result.Err()
	}

	category := types.Category{}
	err := result.Decode(&category)
	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (c *category) FindOneAndDelete(id primitive.ObjectID) (*types.Category, error) {
	result := c.c.FindOneAndUpdate(
		context.Background(),
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"deletedAt": time.Now(),
			"updatedAt": time.Now(),
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	if result.Err() != nil {
		return nil, result.Err()
	}

	category := types.Category{}
	err := result.Decode(&category)
	if err != nil {
		return nil, err
	}

	return &category, nil
}

// TO BE REMOVED

func (c *category) FindByID(id primitive.ObjectID) (*types.Category, error) {
	adminTag := types.Category{}
	filter := bson.M{
		"_id":       id,
		"deletedAt": bson.M{"$exists": false},
	}
	err := c.c.FindOne(context.Background(), filter).Decode(&adminTag)
	if err != nil {
		return nil, e.New(e.EntityNotFound, "Admin tag not found")
	}
	return &adminTag, nil
}

func (c *category) FindTags(name string, page int64) (*types.FindCategoryResult, error) {
	if page < 0 || page == 0 {
		return nil, e.New(e.InvalidPageNumber, "AdminTagMongo FindTags failed")
	}

	var results []*types.Category

	findOptions := options.Find()
	findOptions.SetSkip(viper.GetInt64("page_size") * (page - 1))
	findOptions.SetLimit(viper.GetInt64("page_size"))

	filter := bson.M{
		"name":      primitive.Regex{Pattern: name, Options: "i"},
		"deletedAt": bson.M{"$exists": false},
	}

	cur, err := c.c.Find(context.TODO(), filter, findOptions)
	if err != nil {
		return nil, e.Wrap(err, "AdminTagMongo FindTags failed")
	}

	for cur.Next(context.TODO()) {
		var elem types.Category
		err := cur.Decode(&elem)
		if err != nil {
			return nil, e.Wrap(err, "AdminTagMongo FindTags failed")
		}
		results = append(results, &elem)
	}
	if err := cur.Err(); err != nil {
		return nil, e.Wrap(err, "AdminTagMongo FindTags failed")
	}
	cur.Close(context.TODO())

	// Calculate the total page.
	totalCount, err := c.c.CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, e.Wrap(err, "AdminTagMongo FindTags failed")
	}
	totalPages := util.GetNumberOfPages(int(totalCount), viper.GetInt("page_size"))

	return &types.FindCategoryResult{
		Categories:      results,
		NumberOfResults: int(totalCount),
		TotalPages:      totalPages,
	}, nil
}

func (c *category) GetAll() ([]*types.Category, error) {
	var results []*types.Category

	filter := bson.M{
		"deletedAt": bson.M{"$exists": false},
	}

	cur, err := c.c.Find(context.TODO(), filter)
	if err != nil {
		return nil, e.Wrap(err, "AdminTagMongo GetAll failed")
	}

	for cur.Next(context.TODO()) {
		var elem types.Category
		err := cur.Decode(&elem)
		if err != nil {
			return nil, e.Wrap(err, "AdminTagMongo GetAll failed")
		}
		results = append(results, &elem)
	}
	if err := cur.Err(); err != nil {
		return nil, e.Wrap(err, "AdminTagMongo GetAll failed")
	}
	cur.Close(context.TODO())

	return results, nil
}
