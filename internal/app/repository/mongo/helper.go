package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// A helper function which converts a struct value to a bson.Document.
func toDoc(v interface{}) (doc interface{}, err error) {
	data, err := bson.Marshal(v)
	if err != nil {
		return
	}
	err = bson.Unmarshal(data, &doc)
	return
}

func toObjectIDs(ids []string) ([]primitive.ObjectID, error) {
	objectIDs := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return objectIDs, err
		}
		objectIDs = append(objectIDs, objectID)
	}
	return objectIDs, nil
}

func newFindByIDsPipeline(objectIDs []primitive.ObjectID) []primitive.M {
	return []bson.M{
		{
			"$match": bson.M{
				"_id":       bson.M{"$in": objectIDs},
				"deletedAt": bson.M{"$exists": false},
			},
		},
		{
			"$addFields": bson.M{
				"idOrder": bson.M{"$indexOfArray": bson.A{objectIDs, "$_id"}},
			},
		},
		{
			"$sort": bson.M{"idOrder": 1},
		},
	}
}
