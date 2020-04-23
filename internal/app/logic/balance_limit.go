package logic

import (
	"math"

	"github.com/ic3network/mccs-alpha-api/internal/app/repository/pg"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
)

type balanceLimit struct{}

var BalanceLimit = balanceLimit{}

// IsExceedLimit checks whether or not the account exceeds the max positive or max negative limit.
func (b balanceLimit) IsExceedLimit(accountNumber string, balance float64) (bool, error) {
	limit, err := pg.BalanceLimit.FindByAccountNumber(accountNumber)
	if err != nil {
		return false, err
	}
	if balance < -(math.Abs(limit.MaxNegBal)) || balance > limit.MaxPosBal {
		return true, nil
	}
	return false, nil
}

func (b balanceLimit) FindByAccountNumber(accountNumber string) (*types.BalanceLimit, error) {
	record, err := pg.BalanceLimit.FindByAccountNumber(accountNumber)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (b balanceLimit) GetMaxPosBalance(accountNumber string) (float64, error) {
	balanceLimitRecord, err := pg.BalanceLimit.FindByAccountNumber(accountNumber)
	if err != nil {
		return 0, err
	}
	return balanceLimitRecord.MaxPosBal, nil
}

func (b balanceLimit) GetMaxNegBalance(accountNumber string) (float64, error) {
	balanceLimitRecord, err := pg.BalanceLimit.FindByAccountNumber(accountNumber)
	if err != nil {
		return 0, err
	}
	return math.Abs(balanceLimitRecord.MaxNegBal), nil
}
