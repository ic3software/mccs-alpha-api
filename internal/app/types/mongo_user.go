package types

import (
	"errors"
	"time"

	"github.com/ic3network/mccs-alpha-api/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User is the model representation of an user in the data model.
type User struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	DeletedAt time.Time          `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`

	FirstName string               `json:"firstName,omitempty" bson:"firstName,omitempty"`
	LastName  string               `json:"lastName,omitempty" bson:"lastName,omitempty"`
	Email     string               `json:"email,omitempty" bson:"email,omitempty"`
	Password  string               `json:"password,omitempty" bson:"password,omitempty"`
	Telephone string               `json:"telephone,omitempty" bson:"telephone,omitempty"`
	Entities  []primitive.ObjectID `json:"entities,omitempty" bson:"entities,omitempty"`

	CurrentLoginIP   string    `json:"currentLoginIP,omitempty" bson:"currentLoginIP,omitempty"`
	CurrentLoginDate time.Time `json:"currentLoginDate,omitempty" bson:"currentLoginDate,omitempty"`
	LastLoginIP      string    `json:"lastLoginIP,omitempty" bson:"lastLoginIP,omitempty"`
	LastLoginDate    time.Time `json:"lastLoginDate,omitempty" bson:"lastLoginDate,omitempty"`

	LoginAttempts     int       `json:"loginAttempts,omitempty" bson:"loginAttempts,omitempty"`
	LastLoginFailDate time.Time `json:"lastLoginFailDate,omitempty" bson:"lastLoginFailDate,omitempty"`

	ShowRecentMatchedTags    *bool     `json:"showRecentMatchedTags,omitempty" bson:"showRecentMatchedTags,omitempty"`
	DailyNotification        *bool     `json:"dailyNotification,omitempty" bson:"dailyNotification,omitempty"`
	LastNotificationSentDate time.Time `json:"lastNotificationSentDate,omitempty" bson:"lastNotificationSentDate,omitempty"`
}

func (user *User) Validate() []error {
	errs := []error{}

	if len(user.Email) != 0 {
		errs = append(errs, util.ValidateEmail(user.Email)...)
	}

	if len(user.FirstName) > 100 {
		errs = append(errs, errors.New("First name length cannot exceed 100 characters."))
	}
	if len(user.LastName) > 100 {
		errs = append(errs, errors.New("Last name length cannot exceed 100 characters."))
	}
	if len(user.Telephone) > 25 {
		errs = append(errs, errors.New("Telephone length cannot exceed 25 characters."))
	}
	return errs
}

// Helper types

type SearchUserResult struct {
	Users           []*User
	NumberOfResults int
	TotalPages      int
}
