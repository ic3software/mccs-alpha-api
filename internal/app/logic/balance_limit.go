package logic

import (
	"math"

	"github.com/ic3network/mccs-alpha-api/internal/app/repository/pg"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
)

type balanceLimit struct{}

var BalanceLimit = balanceLimit{}

// IsExceedLimit checks whether or not the account exceeds the max positive or max negative limit.
func (b balanceLimit) IsExceedLimit(id uint, balance float64) (bool, error) {
	limit, err := pg.BalanceLimit.FindByAccountID(id)
	if err != nil {
		return false, err
	}
	if balance < -(math.Abs(limit.MaxNegBal)) || balance > limit.MaxPosBal {
		return true, nil
	}
	return false, nil
}

func (b balanceLimit) GetMaxPosBalance(id uint) (float64, error) {
	balanceLimitRecord, err := pg.BalanceLimit.FindByAccountID(id)
	if err != nil {
		return 0, err
	}
	return balanceLimitRecord.MaxPosBal, nil
}

func (b balanceLimit) GetMaxNegBalance(id uint) (float64, error) {
	balanceLimitRecord, err := pg.BalanceLimit.FindByAccountID(id)
	if err != nil {
		return 0, err
	}
	return math.Abs(balanceLimitRecord.MaxNegBal), nil
}

// TO BE REMOVE

func (b balanceLimit) FindByAccountID(id uint) (*types.BalanceLimit, error) {
	record, err := pg.BalanceLimit.FindByAccountID(id)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (b balanceLimit) FindByEntityID(id string) (*types.BalanceLimit, error) {
	account, err := Account.FindByEntityID(id)
	if err != nil {
		return nil, err
	}
	record, err := pg.BalanceLimit.FindByAccountID(account.ID)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (b balanceLimit) Update(id uint, maxPosBal float64, maxNegBal float64) error {
	err := pg.BalanceLimit.Update(id, maxPosBal, maxNegBal)
	if err != nil {
		return err
	}
	return nil
}
