package pg

import (
	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/e"
	"github.com/jinzhu/gorm"
)

type account struct{}

var Account = &account{}

func (a *account) FindByID(accountID uint) (*types.Account, error) {
	var result types.Account
	err := db.Raw(`
		SELECT A.id, A.balance
		FROM accounts AS A
		WHERE A.id = ?
		LIMIT 1
	`, accountID).Scan(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (a *account) FindByAccountNumber(accountNumber string) (*types.Account, error) {
	var result types.Account
	err := db.Raw(`
		SELECT id, balance
		FROM accounts
		WHERE accounts.account_number = ?
		LIMIT 1
	`, accountNumber).Scan(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (a *account) ifAccountExisted(db *gorm.DB, accountNumber string) bool {
	var result types.Account
	return !db.Raw(`
		SELECT account_number
		FROM accounts
		WHERE accounts.account_number = ?
		LIMIT 1
	`, accountNumber).Scan(&result).RecordNotFound()
}

func (a *account) generateAccountNumber(db *gorm.DB) string {
	accountNumber := goluhn.Generate(16)
	for a.ifAccountExisted(db, accountNumber) {
		accountNumber = goluhn.Generate(16)
	}
	return accountNumber
}

func (a *account) Create() (*types.Account, error) {
	tx := db.Begin()

	accountNumber := a.generateAccountNumber(tx)
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

// TO BE REMOVED

func (a *account) FindByEntityID(entityID string) (*types.Account, error) {
	account := new(types.Account)
	err := db.Where("entity_id = ?", entityID).First(account).Error
	if err != nil {
		return nil, e.New(e.UserNotFound, "user not found")
	}
	return account, nil
}
