package logic

import (
	"errors"

	"github.com/ic3network/mccs-alpha-api/internal/app/repository/mongo"
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/redis"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/util/bcrypt"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type adminUser struct{}

var AdminUser = &adminUser{}

func (a *adminUser) FindByID(id primitive.ObjectID) (*types.AdminUser, error) {
	adminUser, err := mongo.AdminUser.FindByID(id)
	if err != nil {
		return nil, err
	}
	return adminUser, nil
}

func (a *adminUser) FindByIDString(id string) (*types.AdminUser, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	adminUser, err := mongo.AdminUser.FindByID(objectID)
	if err != nil {
		return nil, err
	}
	return adminUser, nil
}

func (a *adminUser) FindByEmail(email string) (*types.AdminUser, error) {
	adminUser, err := mongo.AdminUser.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	return adminUser, nil
}

func (a *adminUser) Login(email string, password string) (*types.AdminUser, error) {
	user, err := mongo.AdminUser.FindByEmail(email)
	if err != nil {
		return &types.AdminUser{}, err
	}

	attempts := redis.GetLoginAttempts(email)
	if attempts >= viper.GetInt("login_attempts.limit") {
		return nil, ErrLoginLocked
	}

	err = bcrypt.CompareHash(user.Password, password)
	if err != nil {
		if attempts+1 >= viper.GetInt("login_attempts.limit") {
			return nil, ErrLoginLocked
		}
		return nil, errors.New("Invalid password.")
	}

	redis.ResetLoginAttempts(email)

	return user, nil
}

func (u *adminUser) IncLoginAttempts(email string) error {
	err := redis.IncLoginAttempts(email)
	if err != nil {
		return err
	}
	return nil
}

func (a *adminUser) UpdateLoginInfo(id primitive.ObjectID, ip string) (*types.LoginInfo, error) {
	info, err := mongo.AdminUser.UpdateLoginInfo(id, ip)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (a *adminUser) ResetPassword(email string, newPassword string) error {
	user, err := mongo.AdminUser.FindByEmail(email)
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.Hash(newPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	err = mongo.AdminUser.UpdatePassword(user)
	if err != nil {
		return err
	}

	return nil
}
