package util

import "go.mongodb.org/mongo-driver/bson/primitive"

func ToObjectIDs(ids []string) []primitive.ObjectID {
	objectIDs := []primitive.ObjectID{}
	for _, id := range ids {
		objectID, _ := primitive.ObjectIDFromHex(id)
		objectIDs = append(objectIDs, objectID)
	}
	return objectIDs
}

func ToIDStrings(ids []primitive.ObjectID) []string {
	idStrings := []string{}
	for _, objID := range ids {
		idStrings = append(idStrings, objID.Hex())
	}
	return idStrings
}

func ContainID(list []primitive.ObjectID, str string) bool {
	for _, item := range list {
		if item == ToObjectID(str) {
			return true
		}
	}
	return false
}

func ToObjectID(id string) primitive.ObjectID {
	objectID, _ := primitive.ObjectIDFromHex(id)
	return objectID
}
