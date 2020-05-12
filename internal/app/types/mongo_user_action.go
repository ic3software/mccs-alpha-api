package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserAction struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"userID,omitempty"`
	Email     string             `bson:"email,omitempty"`
	Action    string             `bson:"action,omitempty"`
	Detail    string             `bson:"detail,omitempty"`
	Category  string             `bson:"category,omitempty"`
	CreatedAt time.Time          `bson:"createdAt,omitempty"`
}
