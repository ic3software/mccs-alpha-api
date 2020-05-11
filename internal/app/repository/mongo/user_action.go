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

// func (u *userAction) Find(c *types.UserActionSearchCriteria, page int64) ([]*types.UserAction, int, error) {
// 	ctx := context.Background()
// 	if page < 0 || page == 0 {
// 		return nil, 0, errors.New("Invalid page number: should start with 1.")
// 	}

// 	var results []*types.UserAction

// 	findOptions := options.Find()
// 	findOptions.SetSkip(viper.GetInt64("page_size") * (page - 1))
// 	findOptions.SetLimit(viper.GetInt64("page_size"))
// 	findOptions.SetSort(bson.M{"createdAt": -1})

// 	filter := bson.M{
// 		"deletedAt": bson.M{"$exists": false},
// 	}
// 	if c.Email != "" {
// 		pattern := c.Email
// 		filter["email"] = primitive.Regex{Pattern: pattern, Options: "i"}
// 	}
// 	if c.Category != "" {
// 		filter["category"] = c.Category
// 	}

// 	// Should not overwrite each others.
// 	if !c.DateFrom.IsZero() || !c.DateTo.IsZero() {
// 		if !c.DateFrom.IsZero() && !c.DateTo.IsZero() {
// 			filter["createdAt"] = bson.M{"$gte": c.DateFrom, "$lte": c.DateTo}
// 		} else if !c.DateFrom.IsZero() {
// 			filter["createdAt"] = bson.M{"$gte": c.DateFrom}
// 		} else {
// 			filter["createdAt"] = bson.M{"$lte": c.DateTo}
// 		}
// 	}

// 	cur, err := u.c.Find(ctx, filter, findOptions)
// 	if err != nil {
// 		return nil, 0, err
// 	}

// 	for cur.Next(ctx) {
// 		var elem types.UserAction
// 		err := cur.Decode(&elem)
// 		if err != nil {
// 			return nil, 0, err
// 		}
// 		results = append(results, &elem)
// 	}
// 	if err := cur.Err(); err != nil {
// 		return nil, 0, err
// 	}
// 	cur.Close(ctx)

// 	// Calculate the total page.
// 	totalCount, err := u.c.CountDocuments(ctx, filter)
// 	if err != nil {
// 		return nil, 0, err
// 	}
// 	totalPages := util.GetNumberOfPages(int(totalCount), viper.GetInt("page_size"))

// 	return results, totalPages, nil
// }
