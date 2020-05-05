package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/util"
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

func (c *category) Create(name string) (*types.Category, error) {
	result := c.c.FindOneAndUpdate(
		context.Background(),
		bson.M{"name": name},
		bson.M{"$setOnInsert": bson.M{"name": name, "createdAt": time.Now()}},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
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

func (c *category) Search(req *types.SearchCategoryReq) (*types.FindCategoryResult, error) {
	var results []*types.Category

	findOptions := options.Find()
	findOptions.SetSkip(int64(req.PageSize * (req.Page - 1)))
	findOptions.SetLimit(int64(req.PageSize))

	filter := bson.M{
		"name":      primitive.Regex{Pattern: "^" + req.Prefix + ".*" + req.Fragment + ".*", Options: "i"},
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
		TotalPages:      util.GetNumberOfPages(int(totalCount), req.PageSize),
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

func (c *category) FindByID(id primitive.ObjectID) (*types.Category, error) {
	category := types.Category{}
	filter := bson.M{
		"_id":       id,
		"deletedAt": bson.M{"$exists": false},
	}
	err := c.c.FindOne(context.Background(), filter).Decode(&category)
	if err != nil {
		return nil, errors.New("Entity not found.")
	}
	return &category, nil
}

func (c *category) FindOneAndUpdate(id primitive.ObjectID, update *types.Category) (*types.Category, error) {
	result := c.c.FindOneAndUpdate(
		context.Background(),
		bson.M{"_id": id},
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
	result := c.c.FindOneAndDelete(
		context.Background(),
		bson.M{"_id": id},
	)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return nil, errors.New("Category does not exist.")
		}
		return nil, result.Err()
	}

	category := types.Category{}
	err := result.Decode(&category)
	if err != nil {
		return nil, err
	}

	return &category, nil
}
