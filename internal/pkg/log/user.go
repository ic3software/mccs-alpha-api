package log

import (
	"fmt"
	"strings"

	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/helper"
	"github.com/ic3network/mccs-alpha-api/util"
)

type user struct{}

var User = user{}

func (us user) Signup(u *types.User, b *types.EntityData) *types.UserAction {
	u.Email = strings.ToLower(u.Email)
	return &types.UserAction{
		UserID: u.ID,
		Email:  u.Email,
		Action: "account created",
		// [entityName] - [firstName] [lastName] - [email]
		ActionDetails: b.EntityName + " - " + u.FirstName + " " + u.LastName + " - " + u.Email,
		Category:      "user",
	}
}

func (us user) LoginSuccess(u *types.User, ip string) *types.UserAction {
	u.Email = strings.ToLower(u.Email)
	return &types.UserAction{
		UserID: u.ID,
		Email:  u.Email,
		Action: "user login successful",
		// [email] - [IP address]
		ActionDetails: u.Email + " - " + ip,
		Category:      "user",
	}
}

func (us user) LoginFailure(u *types.User, ip string) *types.UserAction {
	u.Email = strings.ToLower(u.Email)
	return &types.UserAction{
		UserID: u.ID,
		Email:  u.Email,
		Action: "user login failed",
		// [email] - [IP address]
		ActionDetails: u.Email + " - " + ip,
		Category:      "user",
	}
}

func (us user) LostPassword(u *types.User) *types.UserAction {
	u.Email = strings.ToLower(u.Email)
	return &types.UserAction{
		UserID:        u.ID,
		Email:         u.Email,
		Action:        "sent password reset",
		ActionDetails: u.Email,
		Category:      "user",
	}
}

func (us user) ChangePassword(u *types.User) *types.UserAction {
	u.Email = strings.ToLower(u.Email)
	return &types.UserAction{
		UserID:        u.ID,
		Email:         u.Email,
		Action:        "changed password",
		ActionDetails: u.Email,
		Category:      "user",
	}
}

func (us user) ModifyAccount(
	oldUser *types.User,
	newUser *types.User,
	oldEntity *types.Entity,
	newEntity *types.EntityData,
) *types.UserAction {
	// check for entity
	modifiedFields := util.CheckDiff(oldEntity, newEntity, map[string]bool{"Status": true})
	if !helper.SameTags(newEntity.Offers, oldEntity.Offers) {
		modifiedFields = append(modifiedFields, "offers: "+strings.Join(helper.GetTagNames(oldEntity.Offers), " ")+" -> "+strings.Join(helper.GetTagNames(newEntity.Offers), " "))
	}
	if !helper.SameTags(newEntity.Wants, oldEntity.Wants) {
		modifiedFields = append(modifiedFields, "wants: "+strings.Join(helper.GetTagNames(oldEntity.Wants), " ")+" -> "+strings.Join(helper.GetTagNames(newEntity.Wants), " "))
	}
	// check for user
	modifiedFields = append(modifiedFields, util.CheckDiff(oldUser, newUser, map[string]bool{
		"CurrentLoginIP": true,
		"Password":       true,
		"LastLoginIP":    true,
	})...)
	if len(modifiedFields) == 0 {
		return nil
	}
	return &types.UserAction{
		UserID:        oldUser.ID,
		Email:         newUser.Email,
		Action:        "modified account details",
		ActionDetails: newUser.Email + " - " + strings.Join(modifiedFields, ", "),
		Category:      "user",
	}
}

func (us user) ProposeTransfer(
	proposer *types.User,
	fromEmail string,
	toEmail string,
	amount float64,
	desc string,
) *types.UserAction {
	proposer.Email = strings.ToLower(proposer.Email)
	return &types.UserAction{
		UserID: proposer.ID,
		Email:  proposer.Email,
		Action: "user proposed a transfer",
		// [proposer] - [from] - [to] - [amount] - [desc]
		ActionDetails: proposer.Email + " - " + fromEmail + " - " + toEmail + " - " + fmt.Sprintf("%.2f", amount) + " - " + desc,
		Category:      "user",
	}
}

func (us user) Transfer(
	u *types.User,
	toEmail string,
	amount float64,
	desc string,
) *types.UserAction {
	u.Email = strings.ToLower(u.Email)
	return &types.UserAction{
		UserID: u.ID,
		Email:  u.Email,
		Action: "user transfer",
		// [from] - [to] - [amount] - [desc]
		ActionDetails: u.Email + " - " + toEmail + " - " + fmt.Sprintf("%.2f", amount) + " - " + desc,
		Category:      "user",
	}
}
