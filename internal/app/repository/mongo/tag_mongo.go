package mongo

import (
	"context"
	"time"

	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/e"
	"github.com/ic3network/mccs-alpha-api/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type tag struct {
	c *mongo.Collection
}

var Tag = &tag{}

func (t *tag) Register(db *mongo.Database) {
	t.c = db.Collection("tags")
}

func (t *tag) Find(query *types.SearchTagQuery) (*types.FindTagResult, error) {
	var results []*types.Tag

	findOptions := options.Find()
	findOptions.SetSkip(int64(query.PageSize * (query.Page - 1)))
	findOptions.SetLimit(int64(query.PageSize))

	filter := bson.M{
		"name":      primitive.Regex{Pattern: query.Fragment, Options: "i"},
		"deletedAt": bson.M{"$exists": false},
	}
	cur, err := t.c.Find(context.TODO(), filter, findOptions)
	if err != nil {
		return nil, err
	}

	for cur.Next(context.TODO()) {
		var elem types.Tag
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

	totalCount, err := t.c.CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	return &types.FindTagResult{
		Tags:            results,
		NumberOfResults: int(totalCount),
		TotalPages:      util.GetNumberOfPages(int(totalCount), int(query.PageSize)),
	}, nil
}

func (t *tag) UpdateOffer(name string) (primitive.ObjectID, error) {
	filter := bson.M{"name": name}
	update := bson.M{
		"$set": bson.M{
			"offerAddedAt": time.Now(),
			"updatedAt":    time.Now(),
		},
		"$setOnInsert": bson.M{
			"name":      name,
			"createdAt": time.Now(),
		},
	}
	res := t.c.FindOneAndUpdate(
		context.Background(),
		filter,
		update,
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	)
	if res.Err() != nil {
		return primitive.ObjectID{}, res.Err()
	}

	tag := types.Tag{}
	err := res.Decode(&tag)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return tag.ID, nil
}

func (t *tag) UpdateWant(name string) (primitive.ObjectID, error) {
	filter := bson.M{"name": name}
	update := bson.M{
		"$set": bson.M{
			"wantAddedAt": time.Now(),
			"updatedAt":   time.Now(),
		},
		"$setOnInsert": bson.M{
			"name":      name,
			"createdAt": time.Now(),
		},
	}
	res := t.c.FindOneAndUpdate(
		context.Background(),
		filter,
		update,
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	)
	if res.Err() != nil {
		return primitive.ObjectID{}, res.Err()
	}

	tag := types.Tag{}
	err := res.Decode(&tag)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return tag.ID, nil
}

// TO BE REMOVED

// Create creates a tag record in the table
func (t *tag) Create(name string) (primitive.ObjectID, error) {
	filter := bson.M{"name": name}
	update := bson.M{"$setOnInsert": bson.M{
		"name":      name,
		"createdAt": time.Now(),
	}}
	res, err := t.c.UpdateOne(
		context.Background(),
		filter,
		update,
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return res.UpsertedID.(primitive.ObjectID), nil
}

func (t *tag) FindByName(name string) (*types.Tag, error) {
	tag := types.Tag{}
	filter := bson.M{
		"name":      name,
		"deletedAt": bson.M{"$exists": false},
	}
	err := t.c.FindOne(context.Background(), filter).Decode(&tag)
	if err != nil {
		return nil, e.New(e.EntityNotFound, "Tag not found")
	}
	return &tag, nil
}

func (t *tag) FindByID(id primitive.ObjectID) (*types.Tag, error) {
	tag := types.Tag{}
	filter := bson.M{
		"_id":       id,
		"deletedAt": bson.M{"$exists": false},
	}
	err := t.c.FindOne(context.Background(), filter).Decode(&tag)
	if err != nil {
		return nil, e.New(e.EntityNotFound, "Tag not found")
	}
	return &tag, nil
}

func (t *tag) Rename(tag *types.Tag) error {
	filter := bson.M{"_id": tag.ID}
	update := bson.M{"$set": bson.M{
		"name":      tag.Name,
		"updatedAt": time.Now(),
	}}
	_, err := t.c.UpdateOne(
		context.Background(),
		filter,
		update,
	)
	if err != nil {
		return e.Wrap(err, "TagMongo Update failed")
	}
	return nil
}

func (t *tag) DeleteByID(id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{
		"deletedAt": time.Now(),
		"updatedAt": time.Now(),
	}}
	_, err := t.c.UpdateOne(
		context.Background(),
		filter,
		update,
	)
	if err != nil {
		return e.Wrap(err, "TagMongo DeleteByID failed")
	}
	return nil
}
