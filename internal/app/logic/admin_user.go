package logic

import (
	"errors"
	"time"

	"github.com/ic3network/mccs-alpha-api/internal/app/repository/mongo"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/bcrypt"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/e"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type adminUser struct{}

var AdminUser = &adminUser{}

func (a *adminUser) Login(email string, password string) (*types.AdminUser, error) {
	user, err := mongo.AdminUser.FindByEmail(email)
	if err != nil {
		return &types.AdminUser{}, err
	}

	if a.isUserLockForLogin(user.LastLoginFailDate) {
		return nil, ErrLoginLocked
	}

	err = bcrypt.CompareHash(user.Password, password)
	if err != nil {
		if user.LoginAttempts+1 >= viper.GetInt("login_attempts_limit") {
			return nil, ErrLoginLocked
		}
		return nil, errors.New("Invalid password.")
	}

	return user, nil
}

func (u *adminUser) isUserLockForLogin(lastLoginFailDate time.Time) bool {
	if time.Now().Sub(lastLoginFailDate).Seconds() <= viper.GetFloat64("login_attempts_timeout") {
		return true
	}
	return false
}

func (u *adminUser) UpdateLoginAttempts(email string) error {
	user, err := mongo.AdminUser.FindByEmail(email)
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

	err = mongo.AdminUser.UpdateLoginAttempts(email, attempts, lockUser)
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
