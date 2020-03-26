package util

import "go.mongodb.org/mongo-driver/bson/primitive"

func ToIDStrings(ids []primitive.ObjectID) []string {
	idStrings := []string{}
	for _, objID := range ids {
		idStrings = append(idStrings, objID.Hex())
	}
	return idStrings
}

func ContainID(list []primitive.ObjectID, str string) bool {
	for _, item := range list {
		if item == toObjectID(str) {
			return true
		}
	}
	return false
}

func toObjectID(id string) primitive.ObjectID {
	objectID, _ := primitive.ObjectIDFromHex(id)
	return objectID
}
