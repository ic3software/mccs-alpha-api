package pg

import (
	"fmt"
	"math"

	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/e"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
)

type balanceLimit struct{}

var BalanceLimit = balanceLimit{}

func (b balanceLimit) Create(tx *gorm.DB, accountNumber string) error {
	balance := &types.BalanceLimit{
		AccountNumber: accountNumber,
		MaxNegBal:     viper.GetFloat64("transaction.maxNegBal"),
		MaxPosBal:     viper.GetFloat64("transaction.maxPosBal"),
	}
	err := tx.Create(balance).Error
	if err != nil {
		return err
	}
	return nil
}

func (b balanceLimit) FindByAccountNumber(accountNumber string) (*types.BalanceLimit, error) {
	var result types.BalanceLimit

	err := db.Raw(`
		SELECT max_pos_bal, max_neg_bal
		FROM balance_limits
		WHERE account_number = ?
		LIMIT 1
	`, accountNumber).Scan(&result).Error
	if err != nil {
		fmt.Println(accountNumber)
		return nil, err
	}

	return &result, nil
}

// TO BE REMOVED

func (b balanceLimit) Update(id uint, maxPosBal float64, maxNegBal float64) error {
	if math.Abs(maxNegBal) == 0 {
		maxNegBal = 0
	} else {
		maxNegBal = math.Abs(maxNegBal)
	}

	err := db.
		Model(&types.BalanceLimit{}).
		Where("account_id = ?", id).
		Updates(map[string]interface{}{
			"max_pos_bal": math.Abs(maxPosBal),
			"max_neg_bal": maxNegBal,
		}).Error
	if err != nil {
		return e.Wrap(err, "pg.BalanceLimit.Update failed")
	}
	return nil
}
