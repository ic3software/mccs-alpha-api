package mongo

import (
	"context"
	"time"

	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var UserAction = &userAction{}

type userAction struct {
	c *mongo.Collection
}

func (u *userAction) Register(db *mongo.Database) {
	u.c = db.Collection("userActions")
}

func (u *userAction) Create(a *types.UserAction) (*types.UserAction, error) {
	filter := bson.M{"_id": bson.M{"$exists": false}}
	update := bson.M{
		"userID":    a.UserID,
		"email":     a.Email,
		"action":    a.Action,
		"detail":    a.Detail,
		"category":  a.Category,
		"createdAt": time.Now(),
	}

	result := u.c.FindOneAndUpdate(
		context.Background(),
		filter,
		bson.M{"$set": update},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	)
	if result.Err() != nil {
		return nil, result.Err()
	}

	created := types.UserAction{}
	err := result.Decode(&created)
	if err != nil {
		return nil, err
	}

	return &created, nil
}
