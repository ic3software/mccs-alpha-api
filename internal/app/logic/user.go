package logic

import (
	"errors"
	"time"

	"github.com/ic3network/mccs-alpha-api/internal/app/repository/es"
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/mongo"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/bcrypt"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/e"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type user struct{}

var User = &user{}

func (u *user) Create(email, password string) (primitive.ObjectID, error) {
	_, err := mongo.User.FindByEmail(email)
	if err == nil {
		return primitive.ObjectID{}, e.New(e.EmailExisted, "email existed")
	}

	hashedPassword, err := bcrypt.Hash(password)
	if err != nil {
		return primitive.ObjectID{}, e.Wrap(err, "create user failed")
	}

	userID, err := mongo.User.Create(email, hashedPassword)
	if err != nil {
		return primitive.ObjectID{}, e.Wrap(err, "create user failed")
	}

	err = es.User.Create(userID, email)
	if err != nil {
		return primitive.ObjectID{}, e.Wrap(err, "create user failed")
	}

	return userID, nil
}

func (u *user) AssociateEntity(userID, entityID primitive.ObjectID) error {
	err := mongo.User.AssociateEntity(userID, entityID)
	if err != nil {
		return err
	}
	return nil
}

// OLD CODE

func (u *user) FindByID(id primitive.ObjectID) (*types.User, error) {
	user, err := mongo.User.FindByID(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *user) FindByEmail(email string) (*types.User, error) {
	user, err := mongo.User.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *user) FindByEntityID(id primitive.ObjectID) (*types.User, error) {
	user, err := mongo.User.FindByEntityID(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *user) Login(email string, password string) (*types.User, error) {
	user, err := mongo.User.FindByEmail(email)
	if err != nil {
		return nil, err
	}

	if time.Now().Sub(user.LastLoginFailDate).Seconds() <= viper.GetFloat64("login_attempts_timeout") {
		return nil, e.New(e.AccountLocked, "")
	}

	err = bcrypt.CompareHash(user.Password, password)
	if err != nil {
		return nil, errors.New("Invalid password.")
	}

	return user, nil
}

// UserEmailExists checks if the email exists in the database.
func (u *user) UserEmailExists(email string) bool {
	_, err := mongo.User.FindByEmail(email)
	if err != nil {
		return false
	}
	return true
}

func (u *user) FindUsers(user *types.User, page int64) (*types.FindUserResult, error) {
	ids, numberOfResults, totalPages, err := es.User.Find(user, page)
	if err != nil {
		return nil, e.Wrap(err, "UserService FindUsers failed")
	}
	users, err := mongo.User.FindByIDs(ids)
	if err != nil {
		return nil, e.Wrap(err, "UserService FindUsers failed")
	}
	return &types.FindUserResult{
		Users:           users,
		NumberOfResults: numberOfResults,
		TotalPages:      totalPages,
	}, nil
}

func (u *user) FindByDailyNotification() ([]*types.User, error) {
	users, err := mongo.User.FindByDailyNotification()
	if err != nil {
		return nil, e.Wrap(err, "UserService FindByDailyNotification failed")
	}
	return users, nil
}

// Logout logs out the user.
func (u *user) Logout() error {
	return nil
}

func (u *user) ResetPassword(email string, newPassword string) error {
	user, err := mongo.User.FindByEmail(email)
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.Hash(newPassword)
	if err != nil {
		return e.Wrap(err, "reset password failed")
	}

	user.Password = hashedPassword
	err = mongo.User.UpdatePassword(user)
	if err != nil {
		return e.Wrap(err, "reset password failed")
	}

	return nil
}

func (u *user) UpdateUserInfo(user *types.User) error {
	err := es.User.Update(user)
	if err != nil {
		return e.Wrap(err, "update user info failed")
	}
	err = mongo.User.UpdateUserInfo(user)
	if err != nil {
		return e.Wrap(err, "update user info failed")
	}
	return nil
}

func (u *user) UpdateLastNotificationSentDate(id primitive.ObjectID) error {
	err := mongo.User.UpdateLastNotificationSentDate(id)
	if err != nil {
		return e.Wrap(err, "UserService UpdateLastNotificationSentDate failed")
	}
	return nil
}

func (u *user) AdminUpdateUser(user *types.User) error {
	err := es.User.Update(user)
	if err != nil {
		return e.Wrap(err, "AdminUpdateUser failed")
	}
	err = mongo.User.AdminUpdateUser(user)
	if err != nil {
		return e.Wrap(err, "AdminUpdateUser failed")
	}
	return nil
}

func (u *user) UpdateLoginAttempts(email string) error {
	user, err := mongo.User.FindByEmail(email)
	if err != nil {
		return err
	}

	attempts := user.LoginAttempts
	lockUser := false

	if attempts+1 >= viper.GetInt("login_attempts_limit") {
		attempts = 0
		lockUser = true
	} else {
		attempts++
	}

	err = mongo.User.UpdateLoginAttempts(email, attempts, lockUser)
	if err != nil {
		return err
	}

	return nil
}

func (u *user) UpdateLoginInfo(id primitive.ObjectID, ip string) error {
	loginInfo, err := mongo.User.GetLoginInfo(id)
	if err != nil {
		return e.Wrap(err, "UserService UpdateLoginInfo failed")
	}

	newLoginInfo := &types.LoginInfo{
		CurrentLoginIP: ip,
		LastLoginIP:    loginInfo.CurrentLoginIP,
		LastLoginDate:  loginInfo.CurrentLoginDate,
	}

	err = mongo.User.UpdateLoginInfo(id, newLoginInfo)
	if err != nil {
		return e.Wrap(err, "UserService UpdateLoginInfo failed")
	}
	return nil
}

func (u *user) DeleteByID(id primitive.ObjectID) error {
	err := es.User.Delete(id.Hex())
	if err != nil {
		return e.Wrap(err, "delete user by id failed")
	}
	err = mongo.User.DeleteByID(id)
	if err != nil {
		return e.Wrap(err, "delete user by id failed")
	}
	return nil
}

// APIs

func (u *user) ToggleShowRecentMatchedTags(id primitive.ObjectID) error {
	err := mongo.User.ToggleShowRecentMatchedTags(id)
	if err != nil {
		return e.Wrap(err, "UserService ToggleShowRecentMatchedTags failed")
	}
	return nil
}

func (u *user) AddToFavoriteEntities(uID, entityID primitive.ObjectID) error {
	err := mongo.User.AddToFavoriteEntities(uID, entityID)
	if err != nil {
		return e.Wrap(err, "UserService AddToFavoriteEntities failed")
	}
	return nil
}

func (u *user) RemoveFromFavoriteEntities(uID, entityID primitive.ObjectID) error {
	err := mongo.User.RemoveFromFavoriteEntities(uID, entityID)
	if err != nil {
		return e.Wrap(err, "UserService RemoveFromFavoriteEntities failed")
	}
	return nil
}
