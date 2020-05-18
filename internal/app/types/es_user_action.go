package types

import (
	"time"
)

type UserActionESRecord struct {
	UserID    string    `json:"userID,omitempty"`
	Email     string    `json:"email,omitempty"`
	Action    string    `json:"action,omitempty"`
	Detail    string    `json:"detail,omitempty"`
	Category  string    `json:"category,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

type ESSearchUserActionResult struct {
	UserActions     []*UserActionESRecord
	NumberOfResults int
	TotalPages      int
}
