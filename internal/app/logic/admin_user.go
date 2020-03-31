package logic

import (
	"errors"

	"github.com/ic3network/mccs-alpha-api/internal/app/repository/mongo"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/bcrypt"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/e"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type adminUser struct{}

var AdminUser = &adminUser{}

func (a *adminUser) Login(email string, password string) (*types.AdminUser, error) {
	user, err := mongo.AdminUser.FindByEmail(email)
	if err != nil {
		return &types.AdminUser{}, err
	}

	err = bcrypt.CompareHash(user.Password, password)
	if err != nil {
		return nil, errors.New("Invalid password.")
	}

	return user, nil
}

func (a *adminUser) UpdateLoginInfo(id primitive.ObjectID, ip string) (*types.LoginInfo, error) {
	info, err := mongo.AdminUser.UpdateLoginInfo(id, ip)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// TO BE REMOVED

func (a *adminUser) FindByID(id primitive.ObjectID) (*types.AdminUser, error) {
	adminUser, err := mongo.AdminUser.FindByID(id)
	if err != nil {
		return nil, e.Wrap(err, "service.AdminUser.FindByID failed")
	}
	return adminUser, nil
}

func (a *adminUser) FindByEmail(email string) (*types.AdminUser, error) {
	adminUser, err := mongo.AdminUser.FindByEmail(email)
	if err != nil {
		return nil, e.Wrap(err, "service.AdminUser.FindByEmail failed")
	}
	return adminUser, nil
}
