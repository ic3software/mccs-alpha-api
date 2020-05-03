package logic

import (
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/mongo"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
)

type userAction struct{}

var UserAction = &userAction{}

func (u *userAction) Log(log *types.UserAction) error {
	if log == nil {
		return nil
	}
	err := mongo.UserAction.Log(log)
	if err != nil {
		return err
	}
	return nil
}

func (u *userAction) Find(c *types.UserActionSearchCriteria, page int64) ([]*types.UserAction, int, error) {
	userActions, totalPages, err := mongo.UserAction.Find(c, page)
	if err != nil {
		return nil, 0, err
	}
	return userActions, totalPages, nil
}
