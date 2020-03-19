package seed

import (
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/pg"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
)

var PostgresSQL = postgresSQL{}

type postgresSQL struct{}

func (_ *postgresSQL) CreateAccount() (string, error) {
	account, err := logic.Account.Create()
	if err != nil {
		return "", err
	}
	return account.AccountNumber, nil
}

func (_ *postgresSQL) UpdateBalanceLimits(balanceLimits []types.BalanceLimit) error {
	for _, balanceLimit := range balanceLimits {
		err := pg.DB().Exec(`
			UPDATE balance_limits
			SET max_pos_bal = ?, max_neg_bal = ?
			WHERE id = ?
		`, balanceLimit.MaxPosBal, balanceLimit.MaxNegBal, balanceLimit.AccountID).Error
		if err != nil {
			return err
		}
	}
	return nil
}
