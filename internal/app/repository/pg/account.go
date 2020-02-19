package pg

import (
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/e"
	"github.com/segmentio/ksuid"
)

type account struct{}

var Account = &account{}

func (a *account) Create() (*types.Account, error) {
	tx := db.Begin()

	// TODO
	accountNumber := ksuid.New().String()
	account := &types.Account{AccountNumber: accountNumber, Balance: 0}

	var result types.Account
	err := tx.Create(account).Scan(&result).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = BalanceLimit.Create(tx, account.ID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	return &result, tx.Commit().Error
}

func (a *account) FindByID(accountID uint) (*types.Account, error) {
	var result types.Account
	err := db.Raw(`
	SELECT A.id, A.entity_id, A.balance
	FROM accounts AS A
	WHERE A.id = ?
	LIMIT 1
	`, accountID).Scan(&result).Error
	if err != nil {
		return nil, e.Wrap(err, "pg.Account.FindByID")
	}
	return &result, nil
}

func (a *account) FindByEntityID(entityID string) (*types.Account, error) {
	account := new(types.Account)
	err := db.Where("entity_id = ?", entityID).First(account).Error
	if err != nil {
		return nil, e.New(e.UserNotFound, "user not found")
	}
	return account, nil
}
