package types

import "time"

// LoginInfo is shared by user and admin user model.
type LoginInfo struct {
	CurrentLoginIP   string    `json:"currentLoginIP,omitempty" bson:"currentLoginIP,omitempty"`
	CurrentLoginDate time.Time `json:"currentLoginDate,omitempty" bson:"currentLoginDate,omitempty"`
	LastLoginIP      string    `json:"lastLoginIP,omitempty" bson:"lastLoginIP,omitempty"`
	LastLoginDate    time.Time `json:"lastLoginDate,omitempty" bson:"lastLoginDate,omitempty"`
}
