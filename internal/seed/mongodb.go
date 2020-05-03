package seed

import (
	"context"

	"github.com/ic3network/mccs-alpha-api/internal/app/repository/mongo"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/util/bcrypt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var MongoDB = mongoDB{}

type mongoDB struct{}

func (_ *mongoDB) CreateEntity(entity types.Entity) (primitive.ObjectID, error) {
	res, err := mongo.DB().Collection("entities").InsertOne(context.Background(), entity)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return res.InsertedID.(primitive.ObjectID), nil
}

func (_ *mongoDB) CreateUser(user types.User) (primitive.ObjectID, error) {
	res, err := mongo.DB().Collection("users").InsertOne(context.Background(), user)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return res.InsertedID.(primitive.ObjectID), nil
}

func (_ *mongoDB) AssociateUserWithEntity(userID, entityID primitive.ObjectID) error {
	_, err := mongo.DB().Collection("entities").UpdateOne(context.Background(), bson.M{"_id": entityID}, bson.M{
		"$addToSet": bson.M{"users": userID},
	})
	return err
}

func (_ *mongoDB) CreateAdminUsers(adminUsers []types.AdminUser) error {
	for _, u := range adminUsers {
		hashedPassword, _ := bcrypt.Hash(u.Password)
		u.Password = hashedPassword
		_, err := mongo.DB().Collection("adminUsers").InsertOne(context.Background(), u)
		if err != nil {
			return err
		}
	}
	return nil
}

func (_ *mongoDB) CreateTags(tags []types.Tag) error {
	for _, t := range tags {
		res, err := mongo.DB().Collection("tags").InsertOne(context.Background(), t)
		if err != nil {
			return err
		}
		t.ID = res.InsertedID.(primitive.ObjectID)

		err = ElasticSearch.CreateTag(&t)
		if err != nil {
			return err
		}
	}

	return nil
}

func (_ *mongoDB) CreateCategories(categories []types.Category) error {
	for _, a := range categories {
		_, err := mongo.DB().Collection("categories").InsertOne(context.Background(), a)
		if err != nil {
			return err
		}
	}
	return nil
}
