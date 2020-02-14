package logic

import "go.mongodb.org/mongo-driver/bson/primitive"

func toIDStrings(ids []primitive.ObjectID) []string {
	idStrings := []string{}
	for _, objID := range ids {
		idStrings = append(idStrings, objID.Hex())
	}
	return idStrings
}
