package mongo

import (
	"context"
	"time"

	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/e"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type entity struct {
	c *mongo.Collection
}

var Entity = &entity{}

func (en *entity) Register(db *mongo.Database) {
	en.c = db.Collection("entities")
}

func (b *entity) Create(data *types.Entity) (primitive.ObjectID, error) {
	data.Status = constant.Entity.Pending
	data.CreatedAt = time.Now()
	res, err := b.c.InsertOne(context.Background(), data)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return res.InsertedID.(primitive.ObjectID), nil
}

func (en *entity) AssociateUser(entityID, userID primitive.ObjectID) error {
	filter := bson.M{"_id": entityID}
	update := bson.M{
		"$addToSet": bson.M{"users": userID},
		"$set":      bson.M{"updatedAt": time.Now()},
	}
	_, err := en.c.UpdateOne(
		context.Background(),
		filter,
		update,
	)
	if err != nil {
		return err
	}
	return nil
}

func (b *entity) FindByID(id primitive.ObjectID) (*types.Entity, error) {
	ctx := context.Background()
	entity := types.Entity{}
	filter := bson.M{
		"_id":       id,
		"deletedAt": bson.M{"$exists": false},
	}
	err := b.c.FindOne(ctx, filter).Decode(&entity)
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (b *entity) FindOneAndUpdate(update *types.Entity) (*types.Entity, error) {
	filter := bson.M{"_id": update.ID}
	update.UpdatedAt = time.Now()

	doc, err := toDoc(update)
	if err != nil {
		return nil, err
	}

	result := b.c.FindOneAndUpdate(
		context.Background(),
		filter,
		bson.M{"$set": doc},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	if result.Err() != nil {
		return nil, result.Err()
	}

	entity := types.Entity{}
	err = result.Decode(&entity)
	if err != nil {
		return nil, result.Err()
	}

	return &entity, nil
}

func (b *entity) UpdateTags(id primitive.ObjectID, difference *types.TagDifference) error {
	updates := []bson.M{
		bson.M{"$set": bson.M{"updatedAt": time.Now()}},
	}

	push := bson.M{}
	if len(difference.OffersAdded) != 0 {
		push["offers"] = bson.M{"$each": util.ToTagFields(difference.OffersAdded)}
	}
	if len(difference.WantsAdded) != 0 {
		push["wants"] = bson.M{"$each": util.ToTagFields(difference.WantsAdded)}
	}
	if len(push) != 0 {
		updates = append(updates, bson.M{"$push": push})
	}

	pull := bson.M{}
	if len(difference.OffersRemoved) != 0 {
		pull["offers"] = bson.M{"name": bson.M{"$in": difference.OffersRemoved}}
	}
	if len(difference.WantsRemoved) != 0 {
		pull["wants"] = bson.M{"name": bson.M{"$in": difference.WantsRemoved}}
	}
	if len(pull) != 0 {
		updates = append(updates, bson.M{"$pull": pull})
	}

	var writes []mongo.WriteModel
	for _, update := range updates {
		model := mongo.NewUpdateOneModel().SetFilter(bson.M{"_id": id}).SetUpdate(update)
		writes = append(writes, model)
	}

	_, err := b.c.BulkWrite(context.Background(), writes)
	if err != nil {
		return err
	}
	return nil
}

// OLD CODE

func (b *entity) FindByIDs(ids []string) ([]*types.Entity, error) {
	var results []*types.Entity

	objectIDs, err := toObjectIDs(ids)
	if err != nil {
		return nil, err
	}

	pipeline := newFindByIDsPipeline(objectIDs)
	cur, err := b.c.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	for cur.Next(context.TODO()) {
		var elem types.Entity
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

func (b *entity) UpdateTradingInfo(id primitive.ObjectID, data *types.TradingRegisterData) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{
		"entityName":         data.EntityName,
		"incType":            data.IncType,
		"companyNumber":      data.CompanyNumber,
		"entityPhone":        data.EntityPhone,
		"website":            data.Website,
		"turnover":           data.Turnover,
		"description":        data.Description,
		"locationAddress":    data.LocationAddress,
		"locationCity":       data.LocationCity,
		"locationRegion":     data.LocationRegion,
		"locationPostalCode": data.LocationPostalCode,
		"locationCountry":    data.LocationCountry,
		"status":             constant.Trading.Pending,
		"updatedAt":          time.Now(),
	}}
	_, err := b.c.UpdateOne(
		context.Background(),
		filter,
		update,
	)
	if err != nil {
		return e.Wrap(err, "EntityMongo UpdateTradingInfo failed")
	}
	return nil
}

func (b *entity) SetMemberStartedAt(id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{
		"memberStartedAt": time.Now(),
	}}
	_, err := b.c.UpdateOne(
		context.Background(),
		filter,
		update,
	)
	if err != nil {
		return e.Wrap(err, "EntityMongo SetMemberStartedAt failed")
	}
	return nil
}

func (b *entity) UpdateAllTagsCreatedAt(id primitive.ObjectID, t time.Time) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{
		"offers.$[].createdAt": t,
		"wants.$[].createdAt":  t,
	}}
	_, err := b.c.UpdateMany(context.Background(), filter, update)
	if err != nil {
		return e.Wrap(err, "entityMongo UpdateAllTagsCreatedAt failed")
	}
	return nil
}

func (b *entity) DeleteByID(id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"deletedAt": time.Now()}}
	_, err := b.c.UpdateOne(
		context.Background(),
		filter,
		update,
	)
	if err != nil {
		return e.Wrap(err, "delete entity failed")
	}
	return nil
}

func (b *entity) RenameTag(old string, new string) error {
	err := b.updateOffers(old, new)
	if err != nil {
		return err
	}
	err = b.updateWants(old, new)
	if err != nil {
		return err
	}
	return nil
}

func (b *entity) updateOffers(old string, new string) error {
	filter := bson.M{"offers.name": old}
	update := bson.M{
		"$set": bson.M{
			"offers.$.name": new,
			"updatedAt":     time.Now(),
		},
	}
	_, err := b.c.UpdateMany(context.Background(), filter, update)
	if err != nil {
		return e.Wrap(err, "updateOffers failed")
	}
	return nil
}

func (b *entity) updateWants(old string, new string) error {
	filter := bson.M{"wants.name": old}
	update := bson.M{
		"$set": bson.M{
			"wants.$.name": new,
			"updatedAt":    time.Now(),
		},
	}
	_, err := b.c.UpdateMany(context.Background(), filter, update)
	if err != nil {
		return e.Wrap(err, "updateWants failed")
	}
	return nil
}

func (b *entity) RenameAdminTag(old string, new string) error {
	// Push the new tag tag name.
	filter := bson.M{"adminTags": old}
	update := bson.M{
		"$push": bson.M{"adminTags": new},
		"$set":  bson.M{"updatedAt": time.Now()},
	}
	_, err := b.c.UpdateMany(context.Background(), filter, update)
	if err != nil {
		return e.Wrap(err, "RenameAdminTag failed")
	}
	// Delete the old tag name.
	filter = bson.M{"adminTags": old}
	update = bson.M{
		"$pull": bson.M{"adminTags": old},
		"$set":  bson.M{"updatedAt": time.Now()},
	}
	_, err = b.c.UpdateMany(context.Background(), filter, update)
	if err != nil {
		return e.Wrap(err, "RenameAdminTag failed")
	}
	return nil
}

func (b *entity) DeleteTag(name string) error {
	filter := bson.M{
		"$or": []interface{}{
			bson.M{"offers.name": name},
			bson.M{"wants.name": name},
		},
	}
	update := bson.M{
		"$pull": bson.M{
			"offers": bson.M{"name": name},
			"wants":  bson.M{"name": name},
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}
	_, err := b.c.UpdateMany(
		context.Background(),
		filter,
		update,
	)
	if err != nil {
		return e.Wrap(err, "DeleteTag failed")
	}
	return nil
}

func (b *entity) DeleteAdminTags(name string) error {
	filter := bson.M{
		"$or": []interface{}{
			bson.M{"adminTags": name},
		},
	}
	update := bson.M{
		"$pull": bson.M{
			"adminTags": name,
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}
	_, err := b.c.UpdateMany(
		context.Background(),
		filter,
		update,
	)
	if err != nil {
		return e.Wrap(err, "DeleteAdminTags failed")
	}
	return nil
}
