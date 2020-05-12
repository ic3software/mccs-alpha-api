package logic

import (
	"fmt"
	"strings"

	"github.com/ic3network/mccs-alpha-api/internal/app/repository/es"
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/mongo"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/ic3network/mccs-alpha-api/util/l"
	"go.uber.org/zap"
)

var UserAction = &userAction{}

type userAction struct{}

// POST /signup

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

// POST /login

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

// POST /login

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

func (u *userAction) LostPassword(user *types.User) {
	ua := &types.UserAction{
		UserID: user.ID,
		Email:  user.Email,
		Action: "sent password reset",
		// [email]
		Detail:   user.Email,
		Category: "user",
	}
	u.create(ua)
}

func (u *userAction) ChangePassword(user *types.User) {
	ua := &types.UserAction{
		UserID: user.ID,
		Email:  user.Email,
		Action: "changed password",
		// [email]
		Detail:   user.Email,
		Category: "user",
	}
	u.create(ua)
}

// PATCH /user

func (u *userAction) ModifyUser(origin *types.User, updated *types.User) {
	modifiedFields := util.CheckFieldDiff(origin, updated)
	if len(modifiedFields) == 0 {
		return
	}
	ua := &types.UserAction{
		UserID:   origin.ID,
		Email:    updated.Email,
		Action:   "modified user details",
		Detail:   origin.Email + " - " + updated.Email + ": " + strings.Join(modifiedFields, ", "),
		Category: "user",
	}
	u.create(ua)
}

// PATCH /user/entities/{entityID}

func (u *userAction) ModifyEntity(userID string, origin *types.Entity, updated *types.Entity) {
	user, err := User.FindByStringID(userID)
	if err != nil {
		return
	}
	modifiedFields := util.CheckFieldDiff(origin, updated)
	if len(modifiedFields) == 0 {
		return
	}
	ua := &types.UserAction{
		UserID:   user.ID,
		Email:    user.Email,
		Action:   "user modified entity details",
		Detail:   user.Email + " - " + updated.Email + " - " + strings.Join(modifiedFields, ", "),
		Category: "user",
	}
	u.create(ua)
}

// POST /transfers

func (u *userAction) ProposeTransfer(userID string, req *types.TransferReq) {
	ua := &types.UserAction{
		UserID: util.ToObjectID(userID),
		Email:  req.FromEmail,
		Action: "user proposed a transfer",
		// [proposer] - [from] - [to] - [amount] - [desc]
		Detail:   req.FromEmail + " - " + req.FromAccountNumber + " - " + req.ToAccountNumber + " - " + fmt.Sprintf("%.2f", req.Amount) + " - " + req.Description,
		Category: "user",
	}
	u.create(ua)
}

// PATCH /transfers/{transferID}

func (u *userAction) AcceptTransfer(userID string, j *types.Journal) {
	ua := &types.UserAction{
		UserID: util.ToObjectID(userID),
		Email:  j.FromAccountNumber,
		Action: "user transfer",
		// [from] - [to] - [amount] - [desc]
		Detail:   j.FromAccountNumber + " - " + j.ToAccountNumber + " - " + fmt.Sprintf("%.2f", j.Amount) + " - " + j.Description,
		Category: "user",
	}
	u.create(ua)
}

// POST /admin/login

func (u *userAction) AdminLogin(admin *types.AdminUser, ipAddress string) {
	ua := &types.UserAction{
		UserID: admin.ID,
		Email:  admin.Email,
		Action: "admin user login successful",
		// [email] - [IP address]
		Detail:   admin.Email + " - " + ipAddress,
		Category: "admin",
	}
	u.create(ua)
}

// POST /admin/login

func (u *userAction) AdminLoginFail(email string, ipAddress string) {
	admin, err := mongo.User.FindByEmail(email)
	if err != nil {
		return
	}
	ua := &types.UserAction{
		UserID: admin.ID,
		Email:  admin.Email,
		Action: "admin user login failed",
		// [email] - [IP address]
		Detail:   admin.Email + " - " + ipAddress,
		Category: "admin",
	}
	u.create(ua)
}

// POST /admin/tags

func (u *userAction) AdminCreateTag(userID string, tagName string) {
	admin, err := AdminUser.FindByIDString(userID)
	if err != nil {
		return
	}
	ua := &types.UserAction{
		UserID: admin.ID,
		Email:  admin.Email,
		Action: "admin created new tag",
		// [email] - [Tag]
		Detail:   admin.Email + " - " + tagName,
		Category: "admin",
	}
	u.create(ua)
}

// PATCH /admin/tags/{id}

func (u *userAction) AdminModifyTag(userID string, old string, new string) {
	admin, err := AdminUser.FindByIDString(userID)
	if err != nil {
		return
	}
	ua := &types.UserAction{
		UserID: admin.ID,
		Email:  admin.Email,
		Action: "admin modified a tag",
		// [email] - [Old Tag Name] - [New Tag Name]
		Detail:   admin.Email + " - " + old + " -> " + new,
		Category: "admin",
	}
	u.create(ua)
}

// DELETE /admin/tags/{id}

func (u *userAction) AdminDeleteTag(userID string, tagName string) {
	admin, err := AdminUser.FindByIDString(userID)
	if err != nil {
		return
	}
	ua := &types.UserAction{
		UserID: admin.ID,
		Email:  admin.Email,
		Action: "admin deleted a tag",
		// [email] - [Tag]
		Detail:   admin.Email + " - " + tagName,
		Category: "admin",
	}
	u.create(ua)
}

// POST /admin/categories

func (u *userAction) AdminCreateCategory(userID string, tagName string) {
	admin, err := AdminUser.FindByIDString(userID)
	if err != nil {
		return
	}
	ua := &types.UserAction{
		UserID: admin.ID,
		Email:  admin.Email,
		Action: "admin created a new category",
		// [email] - [Tag]
		Detail:   admin.Email + " - " + tagName,
		Category: "admin",
	}
	u.create(ua)
}

// PATCH /admin/categories/{id}

func (u *userAction) AdminModifyCategory(userID string, old string, new string) {
	admin, err := AdminUser.FindByIDString(userID)
	if err != nil {
		return
	}
	admin.Email = strings.ToLower(admin.Email)
	ua := &types.UserAction{
		UserID: admin.ID,
		Email:  admin.Email,
		// [email] - [Old Tag Name] - [New Tag Name]
		Action:   "admin modified a category",
		Detail:   admin.Email + " - " + old + " -> " + new,
		Category: "admin",
	}
	u.create(ua)
}

func (u *userAction) AdminDeleteCategory(userID string, tagName string) {
	admin, err := AdminUser.FindByIDString(userID)
	if err != nil {
		return
	}
	admin.Email = strings.ToLower(admin.Email)
	ua := &types.UserAction{
		UserID: admin.ID,
		Email:  admin.Email,
		Action: "admin deleted a category",
		// [email] - [Tag]
		Detail:   admin.Email + " - " + tagName,
		Category: "admin",
	}
	u.create(ua)
}

// PATCH /admin/users/{userID}

func (u *userAction) AdminModifyUser(userID string, origin *types.User, updated *types.User) {
	admin, err := AdminUser.FindByIDString(userID)
	if err != nil {
		return
	}
	modifiedFields := util.CheckFieldDiff(origin, updated)
	if len(modifiedFields) == 0 {
		return
	}
	ua := &types.UserAction{
		UserID:   origin.ID,
		Email:    updated.Email,
		Action:   "admin modified user details",
		Detail:   admin.Email + " - " + updated.Email + ": " + strings.Join(modifiedFields, ", "),
		Category: "admin",
	}
	u.create(ua)
}

// DELETE /admin/users/{userID}

func (u *userAction) AdminDeleteUser(userID string, deleted *types.User) {
	admin, err := AdminUser.FindByIDString(userID)
	if err != nil {
		return
	}
	admin.Email = strings.ToLower(admin.Email)
	ua := &types.UserAction{
		UserID: admin.ID,
		Email:  admin.Email,
		Action: "admin deleted a user",
		//
		Detail:   admin.Email + " - " + deleted.Email,
		Category: "admin",
	}
	u.create(ua)
}

// PATCH /admin/entities/{entityID}

func (u *userAction) AdminModifyEntity(userID string, origin *types.Entity, updated *types.Entity) {
	admin, err := AdminUser.FindByIDString(userID)
	if err != nil {
		return
	}
	modifiedFields := util.CheckFieldDiff(origin, updated)
	if len(modifiedFields) == 0 {
		return
	}
	ua := &types.UserAction{
		UserID: admin.ID,
		Email:  admin.Email,
		Action: "admin modified entity details",
		//
		Detail:   admin.Email + " - " + updated.Email + " - " + strings.Join(modifiedFields, ", "),
		Category: "admin",
	}
	u.create(ua)
}

// DELETE /admin/entities/{entityID}

func (u *userAction) AdminDeleteEntity(userID string, deleted *types.Entity) {
	admin, err := AdminUser.FindByIDString(userID)
	if err != nil {
		return
	}
	admin.Email = strings.ToLower(admin.Email)
	ua := &types.UserAction{
		UserID: admin.ID,
		Email:  admin.Email,
		Action: "admin deleted a entity",
		//
		Detail:   admin.Email + " - " + deleted.Email,
		Category: "admin",
	}
	u.create(ua)
}

func (u *userAction) create(ua *types.UserAction) {
	created, err := mongo.UserAction.Create(ua)
	if err != nil {
		l.Logger.Error("userAction.create failed", zap.Error(err))
	}
	err = es.UserAction.Create(created)
	if err != nil {
		l.Logger.Error("userAction.create failed", zap.Error(err))
	}
}

// POST /admin/transfers

func (u *userAction) AdminTransfer(userID string, j *types.Journal) {
	admin, err := AdminUser.FindByIDString(userID)
	if err != nil {
		return
	}
	ua := &types.UserAction{
		UserID: admin.ID,
		Email:  admin.Email,
		Action: "admin transfer for user",
		// admin - [from] -> [to] - [amount]
		Detail:   admin.Email + " - " + j.FromAccountNumber + " -> " + j.ToAccountNumber + " - " + fmt.Sprintf("%.2f", j.Amount) + " - " + j.Description,
		Category: "admin",
	}
	u.create(ua)
}

// GET /admin/log

func (u *userAction) Search(req *types.AdminSearchLogReq) (*types.ESSearchUserActionResult, error) {
	result, err := es.UserAction.Search(req)
	if err != nil {
		return nil, err
	}
	return result, nil
}
