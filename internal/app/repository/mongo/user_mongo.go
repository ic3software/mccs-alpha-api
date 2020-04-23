package mongo

import (
	"context"
	"strings"
	"time"

	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/e"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type user struct {
	c *mongo.Collection
}

var User = &user{}

func (u *user) Register(db *mongo.Database) {
	u.c = db.Collection("users")
}

func (u *user) Create(user *types.User) (primitive.ObjectID, error) {
	user.CreatedAt = time.Now()
	res, err := u.c.InsertOne(context.Background(), user)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return res.InsertedID.(primitive.ObjectID), nil
}

func (u *user) FindByEmail(email string) (*types.User, error) {
	email = strings.ToLower(email)
	if email == "" {
		return &types.User{}, e.New(e.UserNotFound, "Please specify an email address.")
	}

	user := types.User{}
	filter := bson.M{"email": email, "deletedAt": bson.M{"$exists": false}}
	err := u.c.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return nil, e.New(e.UserNotFound, "The specified user could not be found.")
	}

	return &user, nil
}

func (u *user) FindByID(id primitive.ObjectID) (*types.User, error) {
	user := types.User{}
	filter := bson.M{"_id": id, "deletedAt": bson.M{"$exists": false}}
	err := u.c.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *user) FindByIDs(objectIDs []primitive.ObjectID) ([]*types.User, error) {
	var results []*types.User

	pipeline := newFindByIDsPipeline(objectIDs)
	cur, err := u.c.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	for cur.Next(context.TODO()) {
		var elem types.User
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

	return results, nil
}

func (u *user) FindByStringIDs(ids []string) ([]*types.User, error) {
	var results []*types.User

	objectIDs, err := toObjectIDs(ids)
	if err != nil {
		return nil, err
	}

	pipeline := newFindByIDsPipeline(objectIDs)
	cur, err := u.c.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	for cur.Next(context.TODO()) {
		var elem types.User
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

	return results, nil
}

func (u *user) AssociateEntity(userID, entityID primitive.ObjectID) error {
	filter := bson.M{"_id": userID, "deletedAt": bson.M{"$exists": false}}
	update := bson.M{
		"$addToSet": bson.M{"entities": entityID},
		"$set":      bson.M{"updatedAt": time.Now()},
	}
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

func (u *user) FindOneAndUpdate(userID primitive.ObjectID, update *types.User) (*types.User, error) {
	update.Email = strings.ToLower(update.Email)
	update.UpdatedAt = time.Now()

	doc, err := toDoc(update)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": userID, "deletedAt": bson.M{"$exists": false}}

	result := u.c.FindOneAndUpdate(
		context.Background(),
		filter,
		bson.M{"$set": doc},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	if result.Err() != nil {
		return nil, result.Err()
	}

	user := types.User{}
	err = result.Decode(&user)
	if err != nil {
		return nil, result.Err()
	}

	return &user, nil
}

func (u *user) UpdateLoginInfo(id primitive.ObjectID, newLoginIP string) (*types.LoginInfo, error) {
	old := &types.LoginInfo{}
	filter := bson.M{"_id": id, "deletedAt": bson.M{"$exists": false}}
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

	filter = bson.M{"_id": id, "deletedAt": bson.M{"$exists": false}}
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

func (u *user) UpdateLoginAttempts(email string, attempts int, lockUser bool) error {
	filter := bson.M{"email": email, "deletedAt": bson.M{"$exists": false}}
	set := bson.M{
		"loginAttempts": attempts,
	}
	if lockUser {
		set["lastLoginFailDate"] = time.Now()
	}
	update := bson.M{"$set": set}
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

func (u *user) AdminFindOneAndUpdate(userID primitive.ObjectID, update *types.User) (*types.User, error) {
	filter := bson.M{"_id": userID, "deletedAt": bson.M{"$exists": false}}
	update.Email = strings.ToLower(update.Email)
	update.UpdatedAt = time.Now()

	doc, err := toDoc(update)
	if err != nil {
		return nil, err
	}

	result := u.c.FindOneAndUpdate(
		context.Background(),
		filter,
		bson.M{"$set": doc},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	if result.Err() != nil {
		return nil, result.Err()
	}

	user := types.User{}
	err = result.Decode(&user)
	if err != nil {
		return nil, result.Err()
	}

	return &user, nil
}

func (u *user) AdminFindOneAndDelete(id primitive.ObjectID) (*types.User, error) {
	filter := bson.M{"_id": id, "deletedAt": bson.M{"$exists": false}}

	result := u.c.FindOneAndUpdate(
		context.Background(),
		filter,
		bson.M{"$set": bson.M{
			"deletedAt": time.Now(),
			"updatedAt": time.Now(),
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	user := types.User{}
	err := result.Decode(&user)
	if err != nil {
		return nil, result.Err()
	}

	return &user, nil
}

// TO BE REMOVED

func (u *user) FindByEntityID(id primitive.ObjectID) (*types.User, error) {
	user := types.User{}
	filter := bson.M{
		"companyID": id,
		"deletedAt": bson.M{"$exists": false},
	}
	err := u.c.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return nil, e.New(e.UserNotFound, "user not found")
	}
	return &user, nil
}

func (u *user) FindByDailyNotification() ([]*types.User, error) {
	filter := bson.M{
		"dailyNotification": true,
		"deletedAt":         bson.M{"$exists": false},
	}
	projection := bson.M{
		"_id":                      1,
		"email":                    1,
		"companyID":                1,
		"lastNotificationSentDate": 1,
		"dailyNotification":        1,
	}
	findOptions := options.Find()
	findOptions.SetProjection(projection)
	cur, err := u.c.Find(context.TODO(), filter, findOptions)
	if err != nil {
		return nil, err
	}

	var users []*types.User
	for cur.Next(context.TODO()) {
		var elem types.User
		err := cur.Decode(&elem)
		if err != nil {
			return nil, err
		}
		users = append(users, &elem)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	cur.Close(context.TODO())

	return users, nil
}

func (u *user) UpdatePassword(user *types.User) error {
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

func (u *user) UpdateUserInfo(user *types.User) error {
	user.Email = strings.ToLower(user.Email)

	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": bson.M{
		"email":             user.Email,
		"firstName":         user.FirstName,
		"lastName":          user.LastName,
		"telephone":         user.Telephone,
		"dailyNotification": user.DailyNotification,
		"updatedAt":         time.Now(),
	}}
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

func (u *user) GetLoginInfo(id primitive.ObjectID) (*types.LoginInfo, error) {
	loginInfo := &types.LoginInfo{}
	filter := bson.M{"_id": id}
	projection := bson.M{
		"currentLoginIP":   1,
		"currentLoginDate": 1,
		"lastLoginIP":      1,
		"lastLoginDate":    1,
	}
	findOneOptions := options.FindOne()
	findOneOptions.SetProjection(projection)
	err := u.c.FindOne(context.Background(), filter, findOneOptions).Decode(&loginInfo)
	if err != nil {
		return nil, err
	}
	return loginInfo, nil
}

func (u *user) UpdateLastNotificationSentDate(id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{
		"lastNotificationSentDate": time.Now(),
		"updatedAt":                time.Now(),
	}}
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
