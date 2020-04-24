package logic

import (
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/pg"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
)

type account struct{}

var Account = &account{}

func (a *account) Create() (*types.Account, error) {
	account, err := pg.Account.Create()
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (a *account) FindByAccountNumber(accountNumber string) (*types.Account, error) {
	account, err := pg.Account.FindByAccountNumber(accountNumber)
	if err != nil {
		return nil, err
	}
	return account, nil
}

// GET /balance

func (a *account) FindByEntityID(entityID string) (*types.Account, error) {
	entity, err := Entity.FindByStringID(entityID)
	if err != nil {
		return nil, err
	}

	account, err := a.FindByAccountNumber(entity.AccountNumber)
	if err != nil {
		return nil, err
	}

	return account, nil
}
