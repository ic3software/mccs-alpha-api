package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserActionESRecord struct {
	ID        primitive.ObjectID `json:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"userID,omitempty"`
	Email     string             `json:"email,omitempty"`
	Action    string             `json:"action,omitempty"`
	Detail    string             `json:"detail,omitempty"`
	Category  string             `json:"category,omitempty"`
	CreatedAt time.Time          `json:"createdAt,omitempty"`
}
