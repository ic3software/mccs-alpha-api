package logic

import (
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/mongo"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
)

var UserAction = &userAction{}

type userAction struct{}

func (u *userAction) Create(ua *types.UserAction) error {
	if ua == nil {
		return nil
	}
	err := mongo.UserAction.Create(ua)
	if err != nil {
		return err
	}
	return nil
}

// func (u *userAction) Find(c *types.UserActionSearchCriteria, page int64) ([]*types.UserAction, int, error) {
// 	userActions, totalPages, err := mongo.UserAction.Find(c, page)
// 	if err != nil {
// 		return nil, 0, err
// 	}
// 	return userActions, totalPages, nil
// }
