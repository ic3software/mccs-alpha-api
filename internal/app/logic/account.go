package logic

import (
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/pg"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
)

type account struct{}

var Account = &account{}

func (a *account) Create(bID string) error {
	err := pg.Account.Create(bID)
	if err != nil {
		return err
	}
	return nil
}

func (a *account) FindByID(accountID uint) (*types.Account, error) {
	account, err := pg.Account.FindByID(accountID)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (a *account) FindByEntityID(entityID string) (*types.Account, error) {
	account, err := pg.Account.FindByEntityID(entityID)
	if err != nil {
		return nil, err
	}
	return account, nil
}
