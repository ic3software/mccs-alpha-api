package logic

import (
	"errors"
	"time"

	"github.com/ic3network/mccs-alpha-api/internal/app/repository/es"
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/mongo"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/bcrypt"
	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type user struct{}

var User = &user{}

func (u *user) Create(user *types.User) (primitive.ObjectID, error) {
	_, err := mongo.User.FindByEmail(user.Email)
	if err == nil {
		return primitive.ObjectID{}, err
	}

	hashedPassword, err := bcrypt.Hash(user.Password)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	user.Password = hashedPassword

	userID, err := mongo.User.Create(user)
	if err != nil {
		return primitive.ObjectID{}, err
	}

	err = es.User.Create(userID, user)
	if err != nil {
		return primitive.ObjectID{}, err
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

func (u *user) isUserLockForLogin(lastLoginFailDate time.Time) bool {
	if time.Now().Sub(lastLoginFailDate).Seconds() <= viper.GetFloat64("login_attempts_timeout") {
		return true
	}
	return false
}

func (u *user) Login(email string, password string) (*types.User, error) {
	user, err := mongo.User.FindByEmail(email)
	if err != nil {
		return nil, err
	}

	if u.isUserLockForLogin(user.LastLoginFailDate) {
		return nil, ErrLoginLocked
	}

	err = bcrypt.CompareHash(user.Password, password)
	if err != nil {
		if user.LoginAttempts+1 >= viper.GetInt("login_attempts_limit") {
			return nil, ErrLoginLocked
		}
		return nil, errors.New("Invalid password.")
	}

	err = mongo.User.UpdateLoginAttempts(email, 0, false)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *user) FindByID(id primitive.ObjectID) (*types.User, error) {
	user, err := mongo.User.FindByID(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *user) FindByStringID(id string) (*types.User, error) {
	objectID, _ := primitive.ObjectIDFromHex(id)
	user, err := mongo.User.FindByID(objectID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *user) FindByIDs(ids []primitive.ObjectID) ([]*types.User, error) {
	users, err := mongo.User.FindByIDs(ids)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (a *user) UpdateLoginInfo(id primitive.ObjectID, ip string) (*types.LoginInfo, error) {
	info, err := mongo.User.UpdateLoginInfo(id, ip)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (u *user) UpdateLoginAttempts(email string) error {
	user, err := mongo.User.FindByEmail(email)
	if err != nil {
		return err
	}

	if u.isUserLockForLogin(user.LastLoginFailDate) {
		return nil
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

func (u *user) ResetPassword(email string, newPassword string) error {
	user, err := mongo.User.FindByEmail(email)
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.Hash(newPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	err = mongo.User.UpdatePassword(user)
	if err != nil {
		return err
	}

	return nil
}

func (u *user) FindOneAndUpdate(userID primitive.ObjectID, update *types.User) (*types.User, error) {
	err := es.User.Update(userID, update)
	if err != nil {
		return nil, err
	}
	updated, err := mongo.User.FindOneAndUpdate(userID, update)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (u *user) FindEntities(userID primitive.ObjectID) ([]*types.Entity, error) {
	user, err := mongo.User.FindByID(userID)
	if err != nil {
		return nil, err
	}
	entities, err := mongo.Entity.FindByStringIDs(util.ToIDStrings(user.Entities))
	if err != nil {
		return nil, err
	}
	return entities, nil
}

func (u *user) AdminFindOneAndUpdate(userID primitive.ObjectID, update *types.User) (*types.User, error) {
	err := es.User.Update(userID, update)
	if err != nil {
		return nil, err
	}
	updated, err := mongo.User.AdminFindOneAndUpdate(userID, update)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (u *user) AdminFindOneAndDelete(id primitive.ObjectID) (*types.User, error) {
	err := es.User.Delete(id.Hex())
	if err != nil {
		return nil, err
	}
	user, err := mongo.User.AdminFindOneAndDelete(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *user) AdminSearchUser(req *types.AdminSearchUserReqBody) (*types.SearchUserResult, error) {
	result, err := es.User.AdminSearchUser(req)
	if err != nil {
		return nil, err
	}
	users, err := mongo.User.FindByStringIDs(result.UserIDs)
	if err != nil {
		return nil, err
	}
	return &types.SearchUserResult{
		Users:           users,
		NumberOfResults: result.NumberOfResults,
		TotalPages:      result.TotalPages,
	}, nil
}

// TO BE REMOVED

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

// UserEmailExists checks if the email exists in the database.
func (u *user) UserEmailExists(email string) bool {
	_, err := mongo.User.FindByEmail(email)
	if err != nil {
		return false
	}
	return true
}

func (u *user) FindByDailyNotification() ([]*types.User, error) {
	users, err := mongo.User.FindByDailyNotification()
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (u *user) UpdateLastNotificationSentDate(id primitive.ObjectID) error {
	err := mongo.User.UpdateLastNotificationSentDate(id)
	if err != nil {
		return err
	}
	return nil
}
