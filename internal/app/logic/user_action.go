package logic

import (
	"fmt"

	"github.com/ic3network/mccs-alpha-api/internal/app/repository/es"
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/mongo"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/util/l"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

var UserAction = &userAction{}

type userAction struct{}

func (u *userAction) Signup(user *types.User, entity *types.Entity) {
	ua := &types.UserAction{
		UserID: user.ID,
		Email:  user.Email,
		Action: "account created",
		// [EntityName] - [firstName] [lastName] - [email]
		Detail:   entity.EntityName + " - " + user.FirstName + " " + user.LastName + " - " + user.Email,
		Category: "user",
	}
	u.create(ua)
}

func (u *userAction) Login(user *types.User, ipAddress string) {
	ua := &types.UserAction{
		UserID: user.ID,
		Email:  user.Email,
		Action: "user login successful",
		// [email] - [IP address]
		Detail:   user.Email + " - " + ipAddress,
		Category: "user",
	}
	u.create(ua)
}

func (u *userAction) LoginFail(email string, ipAddress string) {
	user, err := mongo.User.FindByEmail(email)
	if err != nil {
		return
	}
	ua := &types.UserAction{
		UserID: user.ID,
		Email:  user.Email,
		Action: "user login failed",
		// [email] - [IP address]
		Detail:   user.Email + " - " + ipAddress,
		Category: "user",
	}
	u.create(ua)
}

func (u *userAction) ChangePassword(user *types.User) {
	ua := &types.UserAction{
		UserID:   user.ID,
		Email:    user.Email,
		Action:   "changed password",
		Detail:   user.Email,
		Category: "user",
	}
	u.create(ua)
}

func (u *userAction) ProposeTransfer(req *types.TransferReq, userID primitive.ObjectID) {
	ua := &types.UserAction{
		UserID: userID,
		Email:  req.FromEmail,
		Action: "user proposed a transfer",
		// [proposer] - [from] - [to] - [amount] - [desc]
		Detail:   req.FromEmail + " - " + req.FromAccountNumber + " - " + req.ToAccountNumber + " - " + fmt.Sprintf("%.2f", req.Amount) + " - " + req.Description,
		Category: "user",
	}
	u.create(ua)
}

func (u *userAction) create(ua *types.UserAction) {
	created, err := mongo.UserAction.Create(ua)
	if err != nil {
		l.Logger.Error("userAction.Login failed", zap.Error(err))
	}
	err = es.UserAction.Create(created)
	if err != nil {
		l.Logger.Error("userAction.Login failed", zap.Error(err))
	}
}

// func (u *userAction) Find(c *types.UserActionSearchCriteria, page int64) ([]*types.UserAction, int, error) {
// 	userActions, totalPages, err := mongo.UserAction.Find(c, page)
// 	if err != nil {
// 		return nil, 0, err
// 	}
// 	return userActions, totalPages, nil
// }
