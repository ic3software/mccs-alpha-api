package util

import "go.mongodb.org/mongo-driver/bson/primitive"

func ToIDStrings(ids []primitive.ObjectID) []string {
	idStrings := []string{}
	for _, objID := range ids {
		idStrings = append(idStrings, objID.Hex())
	}
	return idStrings
}

func ContainID(list []primitive.ObjectID, str primitive.ObjectID) bool {
	for _, item := range list {
		if item == str {
			return true
		}
	}
	return false
}