package mongo

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type adminUser struct {
	c *mongo.Collection
}

var AdminUser = &adminUser{}

func (u *adminUser) Register(db *mongo.Database) {
	u.c = db.Collection("adminUsers")
}

func (u *adminUser) FindByEmail(email string) (*types.AdminUser, error) {
	email = strings.ToLower(email)

	if email == "" {
		return &types.AdminUser{}, errors.New("Please specify an email address.")
	}
	user := types.AdminUser{}
	filter := bson.M{
		"email":     email,
		"deletedAt": bson.M{"$exists": false},
	}
	err := u.c.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return nil, errors.New("The specified admin could not be found.")
	}
	return &user, nil
}

func (u *adminUser) UpdateLoginInfo(id primitive.ObjectID, newLoginIP string) (*types.LoginInfo, error) {
	old := &types.LoginInfo{}
	filter := bson.M{"_id": id}
	projection := bson.M{
		"currentLoginIP":   1,
		"currentLoginDate": 1,
		"lastLoginIP":      1,
		"lastLoginDate":    1,
	}
	findOneOptions := options.FindOne()
	findOneOptions.SetProjection(projection)
	err := u.c.FindOne(context.Background(), filter, findOneOptions).Decode(&old)
	if err != nil {
		return nil, err
	}

	new := &types.LoginInfo{
		CurrentLoginDate: time.Now(),
		CurrentLoginIP:   newLoginIP,
		LastLoginIP:      old.CurrentLoginIP,
		LastLoginDate:    old.CurrentLoginDate,
	}

	filter = bson.M{"_id": id}
	update := bson.M{"$set": bson.M{
		"currentLoginIP":   new.CurrentLoginIP,
		"currentLoginDate": new.CurrentLoginDate,
		"lastLoginIP":      new.LastLoginIP,
		"lastLoginDate":    new.LastLoginDate,
		"updatedAt":        time.Now(),
	}}
	_, err = u.c.UpdateOne(
		context.Background(),
		filter,
		update,
	)
	if err != nil {
		return nil, err
	}

	return new, nil
}

func (u *adminUser) UpdatePassword(user *types.AdminUser) error {
	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": bson.M{"password": user.Password, "updatedAt": time.Now()}}
	_, err := u.c.UpdateOne(
		context.Background(),
		filter,
		update,
	)
	if err != nil {
		return err
	}
	return nil
}

// TO BE REMOVED

func (u *adminUser) FindByID(id primitive.ObjectID) (*types.AdminUser, error) {
	adminUser := types.AdminUser{}
	filter := bson.M{
		"_id":       id,
		"deletedAt": bson.M{"$exists": false},
	}
	err := u.c.FindOne(context.Background(), filter).Decode(&adminUser)
	if err != nil {
		return nil, errors.New("Email address not found.")
	}
	return &adminUser, nil
}
